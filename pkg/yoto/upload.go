package yoto

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Variables rather than constants so tests can shorten them.
var (
	// transcodePollInterval is how often to ask whether a transcode has
	// finished. It is the interval Yoto's own upload examples use.
	transcodePollInterval = 500 * time.Millisecond

	// transcodeTimeout bounds the wait for one. Yoto normalizes and re-encodes
	// in a single pass that normally takes seconds, so something still running
	// this much later is not going to finish; without a bound the CLI would sit
	// there silently forever.
	transcodeTimeout = 15 * time.Minute
)

type UploadURLResponse struct {
	Upload struct {
		// UploadURL is empty when Yoto already holds audio with the sha256 that
		// was asked about. There is nothing to send in that case, and UploadID
		// refers to the copy it already has.
		UploadURL string `json:"uploadUrl"`
		UploadID  string `json:"uploadId"`
	} `json:"upload"`
}

type TranscodeInfo struct {
	Duration int    `json:"duration"`
	FileSize int    `json:"fileSize"`
	Format   string `json:"format"`
	Channels string `json:"channels"`
}

type TranscodeData struct {
	TranscodedSha256 string        `json:"transcodedSha256"`
	Complete         bool          `json:"complete"`
	TranscodedInfo   TranscodeInfo `json:"transcodedInfo"`
}

type TranscodeResponse struct {
	Transcode TranscodeData `json:"transcode"`
}

// FileSHA256 hashes a file the way Yoto names uploads: a SHA-256 digest, base64url
// encoded without padding.
func FileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(hash.Sum(nil)), nil
}

// GetUploadURL asks Yoto for somewhere to put a file.
//
// sha256Hash and filename let Yoto recognise audio it already holds: when it
// does, the response carries no upload URL and the upload can be skipped, which
// is what keeps re-importing a podcast feed from sending every old episode
// again. Both may be empty, which always gets a fresh upload URL back.
func (c *Client) GetUploadURL(sha256Hash string, filename string) (*UploadURLResponse, error) {
	var result UploadURLResponse

	req := c.http.R().SetResult(&result)
	if sha256Hash != "" {
		req.SetQueryParam("sha256", sha256Hash)
	}
	if filename != "" {
		req.SetQueryParam("filename", filename)
	}

	resp, err := req.Get("/media/transcode/audio/uploadUrl")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("api error: %s", resp.String())
	}
	return &result, nil
}

// UploadFile PUTs a local file to a signed upload URL.
func (c *Client) UploadFile(path string, uploadURL string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	// A plain net/http request rather than the client's resty one, for two
	// reasons: the URL is a signed S3 one and must not be sent Yoto's auth
	// header, and this streams the file straight from disk with a known
	// Content-Length. Reading it into memory instead would mean holding the
	// audio of every concurrent upload at once.
	req, err := http.NewRequest(http.MethodPut, uploadURL, file)
	if err != nil {
		return err
	}
	req.ContentLength = info.Size()
	req.Header.Set("Content-Type", "audio/mp3")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed: %s: %s", resp.Status, body)
	}
	return nil
}

// PollTranscode waits for Yoto to finish processing an upload, giving up after
// transcodeTimeout.
func (c *Client) PollTranscode(uploadID string) (*TranscodeData, error) {
	deadline := time.Now().Add(transcodeTimeout)

	for {
		// The API might return the data at root or under "transcode"
		// We'll use a map to handle flexibility
		var result map[string]interface{}
		resp, err := c.http.R().
			SetResult(&result).
			Get(fmt.Sprintf("/media/upload/%s/transcoded", uploadID))

		if err != nil {
			return nil, err
		}
		if resp.IsError() {
			return nil, fmt.Errorf("poll error: %s", resp.String())
		}

		// Try to extract from "transcode" key
		var data TranscodeData
		if t, ok := result["transcode"].(map[string]interface{}); ok {
			// Re-marshal/unmarshal is the easiest way to convert map to struct safely here
			// though less efficient.
			temp, _ := json.Marshal(t)
			json.Unmarshal(temp, &data)
		} else {
			temp, _ := json.Marshal(result)
			json.Unmarshal(temp, &data)
		}

		if data.Complete || data.TranscodedSha256 != "" {
			return &data, nil
		}

		if time.Now().After(deadline) {
			return nil, fmt.Errorf("upload %s was still transcoding after %s", uploadID, transcodeTimeout)
		}
		time.Sleep(transcodePollInterval)
	}
}

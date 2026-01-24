package yoto

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type UploadURLResponse struct {
	Upload struct {
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



func (c *Client) GetUploadURL() (*UploadURLResponse, error) {

	var result UploadURLResponse

	resp, err := c.http.R().

		SetResult(&result).

		Get("/media/transcode/audio/uploadUrl")



	if err != nil {

		return nil, err

	}

	if resp.IsError() {

		return nil, fmt.Errorf("api error: %s", resp.String())

	}

	return &result, nil

}



func (c *Client) UploadFile(path string, uploadURL string) error {

	file, err := os.Open(path)

	if err != nil {

		return err

	}

	defer file.Close()



	stat, err := file.Stat()

	if err != nil {

		return err

	}



	resp, err := c.http.R().

		SetHeader("Content-Type", "audio/mp3").

		SetHeader("Content-Length", fmt.Sprintf("%d", stat.Size())).

		SetBody(file).

		Put(uploadURL)



	if err != nil {

		return err

	}

	if resp.IsError() {

		return fmt.Errorf("upload failed: %s", resp.String())

	}

	return nil

}



func (c *Client) PollTranscode(uploadID string) (*TranscodeData, error) {

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



		time.Sleep(5 * time.Second)

	}

}



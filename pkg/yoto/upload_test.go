package yoto

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// writeTempFile puts content in a file the test owns and returns its path.
func writeTempFile(t *testing.T, name string, content []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
	return path
}

// newTestClient points a client at a mock server. Every response is labelled as
// JSON, which the real API does and which the client needs in order to unmarshal
// anything at all.
func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		handler(w, r)
	}))
	t.Cleanup(server.Close)

	client := NewClient("fake-token", "fake-client-id")
	client.http.SetBaseURL(server.URL)
	return client
}

func TestFileSHA256(t *testing.T) {
	// The expected value is a SHA-256 digest in unpadded base64url, which is
	// how Yoto names uploads: sha256("yotocli test audio\n").
	path := writeTempFile(t, "audio.mp3", []byte("yotocli test audio\n"))

	got, err := FileSHA256(path)
	if err != nil {
		t.Fatalf("FileSHA256 failed: %v", err)
	}

	const want = "bkyMTAoxfyiGts-tEl1dbqEilqowTIwr9Ye2qSq0N9A"
	if got != want {
		t.Errorf("FileSHA256 = %q, want %q", got, want)
	}
	if strings.ContainsAny(got, "+/=") {
		t.Errorf("FileSHA256 = %q, want base64url without padding", got)
	}
}

func TestFileSHA256_MissingFile(t *testing.T) {
	if _, err := FileSHA256(filepath.Join(t.TempDir(), "nope.mp3")); err == nil {
		t.Error("expected an error for a file that does not exist")
	}
}

func TestGetUploadURL(t *testing.T) {
	var gotQuery url.Values
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/media/transcode/audio/uploadUrl" {
			t.Errorf("path = %s", r.URL.Path)
		}
		gotQuery = r.URL.Query()
		fmt.Fprintln(w, `{"upload": {"uploadUrl": "https://s3.example.com/signed", "uploadId": "abc123"}}`)
	})

	upData, err := client.GetUploadURL("somehash", "episode 1.mp3")
	if err != nil {
		t.Fatalf("GetUploadURL failed: %v", err)
	}

	// The hash is what lets Yoto recognise audio it already holds, so it has to
	// actually reach the API.
	if got := gotQuery.Get("sha256"); got != "somehash" {
		t.Errorf("sha256 param = %q, want %q", got, "somehash")
	}
	if got := gotQuery.Get("filename"); got != "episode 1.mp3" {
		t.Errorf("filename param = %q, want %q", got, "episode 1.mp3")
	}
	if upData.Upload.UploadURL != "https://s3.example.com/signed" {
		t.Errorf("UploadURL = %q", upData.Upload.UploadURL)
	}
	if upData.Upload.UploadID != "abc123" {
		t.Errorf("UploadID = %q", upData.Upload.UploadID)
	}
}

func TestGetUploadURL_NoHashOrFilename(t *testing.T) {
	var gotQuery url.Values
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		fmt.Fprintln(w, `{"upload": {"uploadUrl": "https://s3.example.com/signed", "uploadId": "abc123"}}`)
	})

	if _, err := client.GetUploadURL("", ""); err != nil {
		t.Fatalf("GetUploadURL failed: %v", err)
	}

	// Empty values are left off rather than sent blank, which the API would have
	// to interpret as a file whose hash is "".
	if _, ok := gotQuery["sha256"]; ok {
		t.Error("sha256 was sent even though it was empty")
	}
	if _, ok := gotQuery["filename"]; ok {
		t.Error("filename was sent even though it was empty")
	}
}

func TestGetUploadURL_AlreadyHeld(t *testing.T) {
	// Yoto answers with a null upload URL when it recognises the hash: the
	// audio is already there and only the upload ID matters.
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"upload": {"uploadUrl": null, "uploadId": "existing123"}}`)
	})

	upData, err := client.GetUploadURL("somehash", "audio.mp3")
	if err != nil {
		t.Fatalf("GetUploadURL failed: %v", err)
	}

	if upData.Upload.UploadURL != "" {
		t.Errorf("UploadURL = %q, want empty", upData.Upload.UploadURL)
	}
	if upData.Upload.UploadID != "existing123" {
		t.Errorf("UploadID = %q, want %q", upData.Upload.UploadID, "existing123")
	}
}

func TestGetUploadURL_APIError(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintln(w, `{"message": "nope"}`)
	})

	if _, err := client.GetUploadURL("somehash", "audio.mp3"); err == nil {
		t.Error("expected an error for a 403 response")
	}
}

func TestUploadFile(t *testing.T) {
	content := bytes.Repeat([]byte("audio"), 1000)
	path := writeTempFile(t, "audio.mp3", content)

	var gotBody []byte
	var gotLength int64
	var gotChunked bool
	var gotAuth, gotType, gotMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotLength = r.ContentLength
		gotChunked = len(r.TransferEncoding) > 0
		gotAuth = r.Header.Get("Authorization")
		gotType = r.Header.Get("Content-Type")
		gotBody, _ = io.ReadAll(r.Body)
	}))
	defer server.Close()

	client := NewClient("fake-token", "fake-client-id")
	if err := client.UploadFile(path, server.URL); err != nil {
		t.Fatalf("UploadFile failed: %v", err)
	}

	if gotMethod != http.MethodPut {
		t.Errorf("method = %s, want PUT", gotMethod)
	}
	if !bytes.Equal(gotBody, content) {
		t.Errorf("body was %d bytes, want %d", len(gotBody), len(content))
	}
	// The signed URL is S3's, and S3 rejects a chunked PUT, so the length has
	// to be declared up front even though the file is streamed rather than
	// buffered.
	if gotLength != int64(len(content)) {
		t.Errorf("Content-Length = %d, want %d", gotLength, len(content))
	}
	if gotChunked {
		t.Error("request was chunked, want a declared Content-Length")
	}
	// Yoto's bearer token has no business being sent to S3.
	if gotAuth != "" {
		t.Errorf("Authorization header was sent to the upload URL: %q", gotAuth)
	}
	if gotType != "audio/mp3" {
		t.Errorf("Content-Type = %q", gotType)
	}
}

func TestUploadFile_MissingFile(t *testing.T) {
	client := NewClient("fake-token", "fake-client-id")
	if err := client.UploadFile(filepath.Join(t.TempDir(), "nope.mp3"), "https://example.com"); err == nil {
		t.Error("expected an error for a file that does not exist")
	}
}

func TestUploadFile_ServerError(t *testing.T) {
	path := writeTempFile(t, "audio.mp3", []byte("audio"))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "SignatureDoesNotMatch")
	}))
	defer server.Close()

	client := NewClient("fake-token", "fake-client-id")
	err := client.UploadFile(path, server.URL)
	if err == nil {
		t.Fatal("expected an error for a 403 response")
	}
	// The S3 body is the only clue about why a signed upload was refused, so it
	// has to survive into the error.
	if !strings.Contains(err.Error(), "SignatureDoesNotMatch") {
		t.Errorf("error = %v, want it to carry the response body", err)
	}
}

func TestPollTranscode(t *testing.T) {
	restore := shortenPolling(t)
	defer restore()

	calls := 0
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls < 3 {
			fmt.Fprintln(w, `{"transcode": {"complete": false}}`)
			return
		}
		fmt.Fprintln(w, `{"transcode": {"complete": true, "transcodedSha256": "sha1",
			"transcodedInfo": {"duration": 42, "fileSize": 1234, "format": "opus"}}}`)
	})

	data, err := client.PollTranscode("upload1")
	if err != nil {
		t.Fatalf("PollTranscode failed: %v", err)
	}

	if calls != 3 {
		t.Errorf("polled %d times, want 3", calls)
	}
	if data.TranscodedSha256 != "sha1" {
		t.Errorf("TranscodedSha256 = %q", data.TranscodedSha256)
	}
	if data.TranscodedInfo.Duration != 42 || data.TranscodedInfo.FileSize != 1234 {
		t.Errorf("TranscodedInfo = %+v", data.TranscodedInfo)
	}
}

func TestPollTranscode_TimesOut(t *testing.T) {
	restore := shortenPolling(t)
	defer restore()

	// A transcode that never finishes used to hang the CLI forever.
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"transcode": {"complete": false}}`)
	})

	_, err := client.PollTranscode("upload1")
	if err == nil {
		t.Fatal("expected a timeout error")
	}
	if !strings.Contains(err.Error(), "upload1") {
		t.Errorf("error = %v, want it to name the upload", err)
	}
}

// shortenPolling makes the poll loop finish in test time rather than in the
// minutes a real transcode is allowed.
func shortenPolling(t *testing.T) func() {
	t.Helper()
	oldInterval, oldTimeout := transcodePollInterval, transcodeTimeout
	transcodePollInterval = time.Millisecond
	transcodeTimeout = 50 * time.Millisecond
	return func() {
		transcodePollInterval, transcodeTimeout = oldInterval, oldTimeout
	}
}

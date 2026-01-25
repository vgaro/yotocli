package yoto

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
)

const (
	BaseURL = "https://api.yotoplay.com"
)

// Client handles communication with the Yoto API
type Client struct {
	http     *resty.Client
	token    string
	clientID string
}

// NewClient creates a new Yoto API client
func NewClient(token, clientID string) *Client {
	client := resty.New()
	client.SetBaseURL(BaseURL)
	client.SetHeader("User-Agent", "Yoto/2.73 (com.yotoplay.Yoto; build:10405; iOS 17.4.0)")
	
	if token != "" {
		client.SetAuthToken(token)
	}

	return &Client{
		http:     client,
		token:    token,
		clientID: clientID,
	}
}

func (c *Client) ListCards() ([]Card, error) {
	var result LibraryResponse
	resp, err := c.http.R().
		SetResult(&result).
		Get("/card/family/library")

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("api error: %s", resp.String())
	}

	cards := make([]Card, len(result.Cards))
	for i, item := range result.Cards {
		cards[i] = item.Card
	}
	return cards, nil
}

func (c *Client) GetCard(id string) (*Card, error) {
	var result struct {
		Card Card `json:"card"`
	}
	resp, err := c.http.R().
		SetResult(&result).
		Get("/card/" + id)

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("api error: %s", resp.String())
	}

	return &result.Card, nil
}

func (c *Client) DeleteCard(id string) error {
	resp, err := c.http.R().
		Delete("/content/" + id)

	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("api error: %s", resp.String())
	}
	return nil
}

func (c *Client) UpdateCard(id string, card *Card) error {
	// Sanitize icons: Convert https URLs back to yoto:#hash format
	sanitizeCardForUpdate(card)

	// The API for content update seems to use the same endpoint as create (Upsert)
	// We POST to /content, and since the body has cardId, it should update.
	resp, err := c.http.R().
		SetBody(card).
		Post("/content")

	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("api error: %s", resp.String())
	}
	return nil
}

func sanitizeCardForUpdate(card *Card) {
	if card.Content == nil {
		return
	}
	for i := range card.Content.Chapters {
		fixIcon(&card.Content.Chapters[i].Display)
		for j := range card.Content.Chapters[i].Tracks {
			fixIcon(&card.Content.Chapters[i].Tracks[j].Display)
			
			// Ensure Type is set
			if card.Content.Chapters[i].Tracks[j].Type == "" {
				card.Content.Chapters[i].Tracks[j].Type = "audio"
			}
		}
	}
}

func fixIcon(d *Display) {
	if d == nil || d.Icon16x16 == "" {
		return
	}
	if strings.HasPrefix(d.Icon16x16, "http") {
		// Extract last part of path
		parts := strings.Split(d.Icon16x16, "/")
		if len(parts) > 0 {
			hash := parts[len(parts)-1]
			// Sometimes URLs have query params, strip them
			if idx := strings.Index(hash, "?"); idx != -1 {
				hash = hash[:idx]
			}
			// Verify length is 43? Or just try.
			if len(hash) == 43 {
				d.Icon16x16 = "yoto:#" + hash
			}
		}
	}
}

func (c *Client) CreateCard(card *Card) error {
	resp, err := c.http.R().
		SetBody(card).
		Post("/content")

	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("api error: %s", resp.String())
	}
	return nil
}

func (c *Client) DownloadFile(url string, destPath string) error {
	resp, err := c.http.R().
		SetDoNotParseResponse(true).
		Get(url)

	if err != nil {
		return err
	}
	defer resp.RawBody().Close()

	if resp.IsError() {
		return fmt.Errorf("download failed: %s", resp.Status())
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

		_, err = io.Copy(out, resp.RawBody())

		return err

	}

	

	func (c *Client) ListDevices() ([]Device, error) {

		var result DevicesResponse

		resp, err := c.http.R().

			SetResult(&result).

			Get("/device-v2/devices/mine")

	

		if err != nil {

			return nil, err

		}

		if resp.IsError() {

			return nil, fmt.Errorf("api error: %s", resp.String())

		}

		return result.Devices, nil

	}

	

	func (c *Client) GetDeviceStatus(deviceID string) (*DeviceStatus, error) {

		var result struct {

			Status DeviceStatus `json:"status"`

		}

		resp, err := c.http.R().

			SetResult(&result).

			Get("/device-v2/" + deviceID + "/status")

	

		if err != nil {

			return nil, err

		}

		if resp.IsError() {

			return nil, fmt.Errorf("api error: %s", resp.String())

		}

		return &result.Status, nil

	}

	
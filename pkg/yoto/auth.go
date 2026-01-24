package yoto

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

const (
	AuthURL  = "https://login.yotoplay.com/oauth/device/code"
	TokenURL = "https://login.yotoplay.com/oauth/token"
	Audience = "https://api.yotoplay.com"
	Scope    = "openid profile email offline_access"
)

type DeviceAuthResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Error        string `json:"error"`
}

// StartDeviceAuth initiates the device code flow
func (c *Client) StartDeviceAuth() (*DeviceAuthResponse, error) {
	data := url.Values{}
	data.Set("client_id", c.clientID)
	data.Set("scope", Scope)
	data.Set("audience", Audience)

	resp, err := c.http.R().
		SetFormDataFromValues(data).
		SetResult(&DeviceAuthResponse{}).
		Post(AuthURL)

	if err != nil {
		return nil, fmt.Errorf("failed to start auth: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("auth request failed: %s", resp.String())
	}

	return resp.Result().(*DeviceAuthResponse), nil
}

// PollToken polls the token endpoint until the user authorizes or it times out
func (c *Client) PollToken(deviceCode string, interval int) (*TokenResponse, error) {
	// Minimum polling interval
	if interval < 5 {
		interval = 5
	}

	data := url.Values{}
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
	data.Set("device_code", deviceCode)
	data.Set("client_id", c.clientID)

	for {
		resp, err := c.http.R().
			SetFormDataFromValues(data).
			SetResult(&TokenResponse{}).
			Post(TokenURL)

		if err != nil {
			return nil, err // Network error, abort
		}

		// Parse generic error first to handle "authorization_pending"
		var errResp map[string]interface{}
		json.Unmarshal(resp.Body(), &errResp)

		if resp.IsError() {
			errCode, _ := errResp["error"].(string)
			if errCode == "authorization_pending" {
				time.Sleep(time.Duration(interval) * time.Second)
				continue
			}
			if errCode == "slow_down" {
				interval += 5
				time.Sleep(time.Duration(interval) * time.Second)
				continue
			}
			return nil, fmt.Errorf("token error: %v", errResp)
		}

		return resp.Result().(*TokenResponse), nil
	}
}

// RefreshToken exchanges a refresh token for a new access token
func (c *Client) RefreshToken(refreshToken string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", c.clientID)
	data.Set("refresh_token", refreshToken)

	resp, err := c.http.R().
		SetFormDataFromValues(data).
		SetResult(&TokenResponse{}).
		Post(TokenURL)

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, fmt.Errorf("refresh failed: %s", resp.String())
	}

	return resp.Result().(*TokenResponse), nil
}

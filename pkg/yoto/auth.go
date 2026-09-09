package yoto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
)

const (
	AuthorizeURL = "https://login.yotoplay.com/authorize"
	TokenURL     = "https://login.yotoplay.com/oauth/token"
	Audience     = "https://api.yotoplay.com"

	// RedirectURI must be registered verbatim as an Allowed Callback URL
	// in the Yoto developer dashboard (https://dashboard.yoto.dev/).
	RedirectURI = "http://127.0.0.1:8787/callback"

	// CallbackAddr is the loopback address RedirectURI resolves to.
	CallbackAddr = "127.0.0.1:8787"

	// CallbackPath is the path the local callback server serves.
	CallbackPath = "/callback"
)

// OIDCScopes are standard OpenID Connect scopes. They are always available and
// are not listed in the developer dashboard's scope picker.
var OIDCScopes = []string{"openid", "profile", "email"}

// APIScopes are the Yoto API scopes this CLI needs. These are the ones to
// select when registering an application at https://dashboard.yoto.dev/.
// Granting more than this is harmless; granting less breaks some commands.
var APIScopes = []string{
	"offline_access",         // required: issues the refresh token we store
	"family:library:view",    // yoto ls
	"family:library:manage",  // yoto rm
	"user:content:view",      // yoto download, yoto playlist
	"user:content:manage",    // yoto create, add, edit, import
	"user:icons:manage",      // yoto icon
	"family:devices:view",    // yoto status, yoto player
	"family:devices:control", // yoto play/stop/pause, yoto volume
}

// Scope is the space-delimited scope string sent to the authorize endpoint.
var Scope = strings.Join(append(append([]string{}, OIDCScopes...), APIScopes...), " ")

// PKCE holds a code verifier and its derived S256 challenge.
type PKCE struct {
	Verifier  string
	Challenge string
}

// NewPKCE generates a fresh PKCE verifier and its S256 challenge.
func NewPKCE() (*PKCE, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return nil, fmt.Errorf("failed to generate PKCE verifier: %w", err)
	}

	verifier := base64.RawURLEncoding.EncodeToString(raw)
	sum := sha256.Sum256([]byte(verifier))

	return &PKCE{
		Verifier:  verifier,
		Challenge: base64.RawURLEncoding.EncodeToString(sum[:]),
	}, nil
}

// NewState generates an opaque value used to correlate the callback with
// the request this CLI started.
func NewState() (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Error        string `json:"error"`
}

// AuthorizeURLFor builds the browser URL that starts the authorization code
// flow with PKCE.
func (c *Client) AuthorizeURLFor(challenge, state string) string {
	q := url.Values{}
	q.Set("response_type", "code")
	q.Set("client_id", c.clientID)
	q.Set("audience", Audience)
	q.Set("scope", Scope)
	q.Set("redirect_uri", RedirectURI)
	q.Set("code_challenge", challenge)
	q.Set("code_challenge_method", "S256")
	q.Set("state", state)

	return AuthorizeURL + "?" + q.Encode()
}

// ExchangeCode trades an authorization code plus its PKCE verifier for tokens.
func (c *Client) ExchangeCode(code, verifier string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", c.clientID)
	data.Set("code", code)
	data.Set("code_verifier", verifier)
	data.Set("redirect_uri", RedirectURI)

	resp, err := c.http.R().
		SetFormDataFromValues(data).
		SetResult(&TokenResponse{}).
		Post(TokenURL)

	if err != nil {
		return nil, fmt.Errorf("failed to exchange authorization code: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("code exchange failed: %s", resp.String())
	}

	return resp.Result().(*TokenResponse), nil
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

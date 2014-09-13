// client defines a client struct to contain platform authentication information
// and is used by both go-soundcloud-api and go-youtube-api
//
// Platform ids are: Soundcloud - 0, Youtube - 1

package oauth

import (
	"encoding/json"
	"fmt"
	"apperror"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	ID          string
	Secret      string
	RedirectURI string
	Platform    int
	Token       Token
	Client      *http.Client
}

type Token struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	ExpiresAt    time.Time
	Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token"`
	Platform     int
}

func NewClient(id, secret, rURI string, pID int) *Client {
	return &Client{
		ID:          id,
		Secret:      secret,
		RedirectURI: rURI,
		Platform:    pID,
		Client:      &http.Client{},
	}
}

// Return a user sign in url where the user may grant access to us. Accepts the base url, and params
func GetSignIn(uri string, params map[string]string) string {
	v := url.Values{}

	for key, val := range params {
		v.Set(key, val)
	}

	return uri + "?" + v.Encode()
}

func Auth(c *Client, code, uri string) error {
	v := url.Values{}
	v.Set("client_id", c.ID)
	v.Set("client_secret", c.Secret)
	v.Set("redirect_uri", c.RedirectURI)
	v.Set("grant_type", "authorization_code")
	v.Set("code", code)

	params := strings.NewReader(v.Encode())

	req, err := http.NewRequest("POST", uri, params)
	if err != nil {
		return apperror.Err{err, "Can't make request", 500}
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.Client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return apperror.Err{err, "Error with request", resp.StatusCode}
	}

	// Copy the access token from the response SoundCloud JSON to our YoutubeClient's Token struct
	json.NewDecoder(resp.Body).Decode(&c.Token)

	if c.Token.ExpiresIn < 1 {
		c.Token.ExpiresAt = SetExpire(3153600000)
	} else {
		c.Token.ExpiresAt = SetExpire(c.Token.ExpiresIn)
	}

	return nil
}

// Refresh an expired auth token with a refresh token
func Refresh(c *Client, uri string) error {
	v := url.Values{}
	v.Set("client_id", c.ID)
	v.Set("client_secret", c.Secret)
	v.Set("refresh_token", c.Token.RefreshToken)
	v.Set("grant_type", "refresh_token")

	params := strings.NewReader(v.Encode())

	req, err := http.NewRequest("POST", uri, params)
	if err != nil {
		return apperror.Err{err, "Can't make request", 500}
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.Client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return apperror.Err{err, "Error with request", resp.StatusCode}
	}
	// Copy the access token from the response SoundCloud JSON to our YoutubeClient's Token struct
	json.NewDecoder(resp.Body).Decode(&c.Token)

	if c.Token.ExpiresIn < 1 {
		c.Token.ExpiresAt = SetExpire(3153600000)
	} else {
		c.Token.ExpiresAt = SetExpire(c.Token.ExpiresIn)
	}

	return nil
}

func SetExpire(expiresIn int) time.Time {
	t := time.Now()
	fmt.Println("time", t)
	return t.Add(time.Duration(expiresIn) * time.Second)
}

func GetExpire(expiresDate time.Time) bool {
	t := time.Now()
	return expiresDate.Before(t)
}

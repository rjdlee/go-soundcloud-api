//Dependencies: apperror, platform/oauth

package soundcloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"apperror"
	"platform/oauth"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

// Credentials for Soundcloud API: https://soundcloud.com/you/apps/
const (
	cID       = "YOUR_CLIENT_ID"
	cSecret   = "YOUR_CLIENT_SECRET"
	cRedirect = "YOUR_REDIRECT_LINK"
)

type Track struct {
	Title         string `json: "title"`
	Description   string `json: "description"`
	Genre         string `json: "genre"`
	Permalink_url string `json: "permalink_url"`
}

// Return a user sign in url where the user may grant access to us
func GetSignIn() string {
	params := map[string]string{
		"client_id":     cID,
		"redirect_uri":  cRedirect,
		"response_type": "code",
		"scope":         "non-expiring",
	}
	return oauth.GetSignIn("https://soundcloud.com/connect", params)
}

// Authenticate and receive a access token using the client code after they provide access to us
func Auth(code string) (*oauth.Client, error) {
	c := oauth.NewClient(cID, cSecret, cRedirect, 0)
	err := oauth.Auth(c, code, "https://api.soundcloud.com/oauth2/token")
	if err != nil {
		return c, err
	}

	return c, nil
}

// Upload a track to Soundcloud
func Upload(path string, c *oauth.Client) (*map[string]interface{}, error) {
	params := map[string]string{
		"oauth_token":    c.Token.AccessToken,
		"track[title]":   "Test Track",
		"track[sharing]": "public",
	}

	soundFile, err := os.Open(path)
	if err != nil {
		return nil, apperror.Err{err, "Unable to open file", 500}
	}
	defer soundFile.Close()

	imageFile, err := os.Open("temppics/young.jpg")
	if err != nil {
		return nil, apperror.Err{err, "Unable to open file", 500}
	}
	defer imageFile.Close()

	// Create a buffer containg a form with information about the track and the track itself
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Attach the audio file to the form
	soundPart, err := writer.CreateFormFile("track[asset_data]", filepath.Base(path))
	if err != nil {
		return nil, apperror.Err{err, "Unable to attach file to form", 500}
	}
	_, err = io.Copy(soundPart, soundFile)

	// Attach the picture file to the form
	imagePart, err := writer.CreateFormFile("track[artwork_data]", filepath.Base("temppics/young.jpg"))
	if err != nil {
		return nil, apperror.Err{err, "Unable to attach file to form", 500}
	}
	_, err = io.Copy(imagePart, imageFile)

	// Iterate through each parameter and add it to a field in the form
	for key, val := range params {
		err := writer.WriteField(key, val)
		if err != nil {
			return nil, apperror.Err{err, "Unable to write field", 500}
		}
	}
	err = writer.Close()
	if err != nil {
		return nil, apperror.Err{err, "Unable to close writer", 500}
	}

	req, err := http.NewRequest("POST", "https://api.soundcloud.com/tracks.json", body)
	if err != nil {
		return nil, apperror.Err{err, "Can't make request", 500}
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.Client.Do(req)
	fmt.Println(resp, err, c.Token.AccessToken)
	if err != nil {
		return nil, apperror.Err{err, "Error with request", resp.StatusCode}
	}

	var content map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&content)

	return &content, nil
}

// Retrieve all of a user's tracks and information about them
func GetTracks(c *oauth.Client) ([]Track, error) {
	// Convert our string into a *url.URL
	u, err := url.Parse("https://api.soundcloud.com/me/tracks.json")
	if err != nil {
		return nil, apperror.Err{err, "Could not parse url", 500}
	}

	q := u.Query()
	q.Set("oauth_token", c.Token.AccessToken)

	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, apperror.Err{err, "Can't make request", 500}
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, apperror.Err{err, "Error with request", resp.StatusCode}
	}

	var content []Track
	err = json.NewDecoder(resp.Body).Decode(&content)
	if err != nil {
		return nil, apperror.Err{err, "Unable to decode response JSON", 500}
	}

	return content, nil
}

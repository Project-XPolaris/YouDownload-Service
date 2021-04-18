package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/projectxpolaris/youdownload-server/config"
	"net/http"
)

var DefaultAuthClient = AuthClient{}

type AuthClient struct {
}

func (c *AuthClient) GetUrl(path string) string {
	return fmt.Sprintf("%s/%s", config.Instance.AuthUrl, path)
}

func (c *AuthClient) makePOSTRequest(url string, data interface{}, responseBody interface{}) error {
	rawData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	client := http.Client{}
	request, err := http.NewRequest("POST", c.GetUrl(url), bytes.NewBuffer(rawData))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	err = json.NewDecoder(response.Body).Decode(&responseBody)
	return err
}

func (c *AuthClient) makeGETRequest(url string, responseBody interface{}) error {
	response, err := http.Get(c.GetUrl(url))
	err = json.NewDecoder(response.Body).Decode(&responseBody)
	return err
}

type AuthResponse struct {
	Success  bool   `json:"success,omitempty"`
	Username string `json:"username,omitempty"`
	Uid      string `json:"uid,omitempty"`
}

func (c *AuthClient) CheckAuth(token string) (*AuthResponse, error) {
	var responseBody AuthResponse
	err := c.makeGETRequest(fmt.Sprintf("%s?token=%s", "/user/auth", token), &responseBody)
	if err != nil {
		return nil, err
	}
	return &responseBody, nil
}

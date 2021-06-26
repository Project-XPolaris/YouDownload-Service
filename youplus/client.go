package youplus

import (
	"fmt"
	"github.com/allentom/haruka"
	"github.com/go-resty/resty/v2"
)

var DefaultAuthClient = AuthClient{
	client: resty.New(),
}

type AuthClient struct {
	client  *resty.Client
	baseUrl string
}

func (c *AuthClient) Init(baseUrl string) {
	c.baseUrl = baseUrl
}
func (c *AuthClient) GetUrl(path string) string {
	return fmt.Sprintf("%s%s", c.baseUrl, path)
}

type AuthResponse struct {
	Success  bool   `json:"success,omitempty"`
	Username string `json:"username,omitempty"`
	Uid      string `json:"uid,omitempty"`
}

func (c *AuthClient) CheckAuth(token string) (*AuthResponse, error) {
	var responseBody AuthResponse
	_, err := c.client.R().
		SetResult(&responseBody).
		SetQueryParam("token", token).
		Get(fmt.Sprintf(c.GetUrl("/user/auth")))
	if err != nil {
		return nil, err
	}
	return &responseBody, nil
}

type UserAuthResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token"`
	Uid     string `json:"uid"`
}

func (c *AuthClient) FetchUserAuth(username string, password string) (*UserAuthResponse, error) {
	var responseBody UserAuthResponse
	_, err := c.client.R().SetBody(haruka.JSON{
		"username": username,
		"password": password,
	}).SetResult(&responseBody).Post(c.GetUrl("/user/auth"))
	return &responseBody, err
}

type InfoResponse struct {
	Success bool `json:"success"`
}

func (c *AuthClient) FetchInfo() (*InfoResponse, error) {
	var responseBody InfoResponse
	_, err := c.client.R().SetResult(&responseBody).Get(c.GetUrl("/info"))
	return &responseBody, err
}

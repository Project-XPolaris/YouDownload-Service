package youplus

import (
	"fmt"
	"github.com/go-resty/resty/v2"
)

var DefaultClient = Client{
	client: resty.New(),
}

type Client struct {
	client  *resty.Client
	baseUrl string
}

// Init client
func (c *Client) Init(baseUrl string) {
	c.baseUrl = baseUrl
}
func (c *Client) GetUrl(path string) string {
	return fmt.Sprintf("%s%s", c.baseUrl, path)
}

type AuthResponse struct {
	Success  bool   `json:"success,omitempty"`
	Username string `json:"username,omitempty"`
	Uid      string `json:"uid,omitempty"`
}

// CheckAuth get user info by token
func (c *Client) CheckAuth(token string) (*AuthResponse, error) {
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

// FetchUserAuth fetch token with username and password
func (c *Client) FetchUserAuth(username string, password string) (*UserAuthResponse, error) {
	var responseBody UserAuthResponse
	_, err := c.client.R().SetBody(map[string]interface{}{
		"username": username,
		"password": password,
	}).SetResult(&responseBody).Post(c.GetUrl("/user/auth"))
	return &responseBody, err
}

type InfoResponse struct {
	Success bool `json:"success"`
}

// FetchInfo get service info
func (c *Client) FetchInfo() (*InfoResponse, error) {
	var responseBody InfoResponse
	_, err := c.client.R().SetResult(&responseBody).Get(c.GetUrl("/info"))
	return &responseBody, err
}

type GetRealPathResponseBody struct {
	Path string `json:"path"`
}

// GetRealPath get realpath by youplus path
func (c *Client) GetRealPath(target string, token string) (string, error) {
	var responseBody GetRealPathResponseBody
	_, err := c.client.R().
		SetQueryParam("target", target).
		SetHeader("Authorization", token).
		SetResult(&responseBody).
		Get(c.baseUrl + "/path/realpath")
	return responseBody.Path, err
}

type GetInfoResponseBody struct {
	Name    string `json:"name"`
	Success bool   `json:"success"`
}

func (c *Client) GetInfo() (*GetInfoResponseBody, error) {
	var responseBody GetInfoResponseBody
	_, err := c.client.R().
		SetResult(&responseBody).
		Get(c.baseUrl + "/info")
	return &responseBody, err
}

type ReadDirItem struct {
	RealPath string `json:"realPath"`
	Path     string `json:"path"`
	Type     string `json:"type"`
}

// ReadDir readdir with youplus path
func (c *Client) ReadDir(target string, token string) ([]ReadDirItem, error) {
	var responseBody []ReadDirItem
	_, err := c.client.R().
		SetQueryParam("target", target).
		SetHeader("Authorization", token).
		SetResult(&responseBody).
		Get(c.baseUrl + "/path/readdir")
	return responseBody, err
}

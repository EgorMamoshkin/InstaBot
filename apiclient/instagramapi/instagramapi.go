package instagramapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/EgorMamoshkin/InstaBot/lib/er"
	"io"
	"net/http"
	"net/url"
	"path"
)

type Client struct {
	host   string
	cfg    ClientCfg
	client http.Client
}

type ClientCfg struct {
	appID       string
	appSecret   string
	redirectURI string
}

func New(host string, appID string, appSecret string, redirectURI string) *Client {
	cgf := ClientCfg{
		appID:       appID,
		appSecret:   appSecret,
		redirectURI: redirectURI,
	}
	return &Client{
		host:   host,
		cfg:    cgf,
		client: http.Client{},
	}
}

func (c *Client) GetRedirectURI() string {
	return c.cfg.redirectURI
}

func (c *Client) GetAppID() string {
	return c.cfg.appID
}

func (c *Client) GetAPIHost() string {
	return c.host
}

func (c *Client) GetAccessToken(reqToken string) (*User, error) {
	qr := url.Values{}
	qr.Add("client_id", c.cfg.appID)
	qr.Add("client_secret", c.cfg.appSecret)
	qr.Add("grant_type", "authorization_code")
	qr.Add("redirect_uri", c.cfg.redirectURI)
	qr.Add("code", reqToken)

	apiPath := path.Join("oauth", "access_token")

	data, err := c.doRequest(qr, apiPath)
	if err != nil {
		return nil, er.Wrap("can't get access token: %s", err)
	}

	var userToken User

	err = json.Unmarshal(data, &userToken)
	if err != nil {
		return nil, er.Wrap("can't unmarshal response: %s", err)
	}

	if userToken.UserID == 0 || userToken.Token == "" {
		return nil, errors.New(fmt.Sprintf("instagramapi refused to receive a token: %s", string(data)))
	}

	return &userToken, nil
}

func (c *Client) doRequest(query url.Values, path string) ([]byte, error) {
	const reqError = "request failed"

	u := url.URL{
		Scheme: "https",
		Host:   c.host,
		Path:   path,
	}

	resp, err := http.PostForm(u.String(), query)
	if err != nil {
		return nil, er.Wrap(reqError, err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, er.Wrap(reqError, err)
	}

	return body, nil
}

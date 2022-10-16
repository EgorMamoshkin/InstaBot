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
	"strconv"
)

const (
	getMediaAPIPath = "media"
)

type Client struct {
	apiHost  string
	authHost string
	cfg      ClientCfg
	client   http.Client
}

type ClientCfg struct {
	appID       string
	appSecret   string
	redirectURI string
}

func New(apiHost string, authHost string, appID string, appSecret string, redirectURI string) *Client {
	cgf := ClientCfg{
		appID:       appID,
		appSecret:   appSecret,
		redirectURI: redirectURI,
	}
	return &Client{
		apiHost:  apiHost,
		authHost: authHost,
		cfg:      cgf,
		client:   http.Client{},
	}
}

func (c *Client) GetRedirectURI() string {
	return c.cfg.redirectURI
}

func (c *Client) GetAppID() string {
	return c.cfg.appID
}

func (c *Client) GetAPIHost() string {
	return c.authHost
}

func (c *Client) ExchangeToAccessToken(reqToken string) (*User, error) {
	qr := url.Values{}
	qr.Add("client_id", c.cfg.appID)
	qr.Add("client_secret", c.cfg.appSecret)
	qr.Add("grant_type", "authorization_code")
	qr.Add("redirect_uri", c.cfg.redirectURI)
	qr.Add("code", reqToken)

	apiPath := path.Join("oauth", "access_token")

	data, err := c.doAuthRequest(qr, apiPath)
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

func (c *Client) GetPosts(instUser *User) (*UserMedia, error) {
	qr := url.Values{}
	qr.Add("access_token", instUser.Token)
	qr.Add("fields", "id, caption, media_type, media_url, username, timestamp, children")

	apiPath := path.Join("v15.0", strconv.Itoa(instUser.UserID), getMediaAPIPath)

	data, err := c.doRequest(qr, apiPath)
	if err != nil {
		return nil, er.Wrap("can't get posts", err)
	}

	var mediaData UserMedia

	err = json.Unmarshal(data, &mediaData)
	if err != nil {
		return nil, er.Wrap("can't unmarshal posts", err)
	}

	return &mediaData, nil
}

func (c *Client) CarouselElements(id string, instUser *User) (*UserMedia, error) {
	qr := url.Values{}
	qr.Add("access_token", instUser.Token)
	qr.Add("fields", "id, media_type, media_url")

	var carousel UserMedia

	data, err := c.doRequest(qr, path.Join(id, "children"))
	if err != nil {
		return nil, er.Wrap("can't get media", err)
	}

	err = json.Unmarshal(data, &carousel)
	if err != nil {
		return nil, er.Wrap("can't unmarshal media", err)
	}

	return &carousel, nil
}

func (c *Client) doAuthRequest(query url.Values, apiPath string) ([]byte, error) {
	const reqError = "request failed"

	u := url.URL{
		Scheme: "https",
		Host:   c.authHost,
		Path:   apiPath,
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

func (c *Client) doRequest(query url.Values, apiPath string) ([]byte, error) {
	const reqError = "request failed"

	u := url.URL{
		Scheme: "https",
		Host:   c.apiHost,
		Path:   apiPath,
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, er.Wrap("can't create request: ", err)
	}

	req.URL.RawQuery = query.Encode()

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, er.Wrap(reqError, err)
	}

	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, er.Wrap(reqError, err)
	}

	return body, nil
}

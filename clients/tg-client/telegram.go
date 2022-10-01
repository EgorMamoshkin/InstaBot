package tg_client

import (
	"InstaBot/lib/er"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
)

type Client struct {
	host     string
	basePath string
	client   http.Client
}

const (
	getUpdatesMethod  = "getUpdates"
	sendMessageMethod = "sendMessage"
)

func New(host string, token string) Client {
	return Client{
		host:     "host",
		basePath: newBasePath(token),
		client:   http.Client{},
	}
}

func newBasePath(token string) string {
	return "bot" + token
}

func (c *Client) Updates(offset int, limit int) ([]Updates, error) {
	const getUpdError = "Getting updates failed"
	qr := url.Values{}
	qr.Add("offset", strconv.Itoa(offset))
	qr.Add("limit", strconv.Itoa(limit))

	data, err := c.doRequest(getUpdatesMethod, qr)
	if err != nil {
		return nil, er.Wrap(getUpdError, err)
	}

	var resp UpdatesResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, er.Wrap(getUpdError, err)
	}
	return resp.Result, nil
}

func (c *Client) SendMessage(chatID int, text string) error {
	qr := url.Values{}
	qr.Add("chat_id", strconv.Itoa(chatID))
	qr.Add("text", text)

	_, err := c.doRequest(sendMessageMethod, qr)
	if err != nil {
		return er.Wrap("sending the message failed", err)
	}
	return nil
}

func (c *Client) doRequest(method string, query url.Values) ([]byte, error) {
	const reqError = "request failed"
	u := url.URL{
		Scheme: "https",
		Host:   c.host,
		Path:   path.Join(c.basePath, method),
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, er.Wrap(reqError, err)
	}

	req.URL.RawQuery = query.Encode()
	resp, err := c.client.Do(req)
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

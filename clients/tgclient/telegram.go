package tgclient

import (
	"InstaBot/lib/er"
	"encoding/json"
	"errors"
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
	getUpdatesMethod     = "getUpdates"
	sendMessageMethod    = "sendMessage"
	sendPhotoMethod      = "sendPhoto"
	sendVideoMethod      = "sendVideo"
	sendMediaGroupMethod = "sendMediaGroup"
)

func New(host string, token string) Client {
	return Client{
		host:     host,
		basePath: newBasePath(token),
		client:   http.Client{},
	}
}

func newBasePath(token string) string {
	return "bot" + token
}

func (c *Client) Updates(offset int, limit int) ([]Update, error) {
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

func (c *Client) SendPhoto(chatID int, photoURL string, caption string) error {
	qr := url.Values{}
	qr.Add("chat_id", strconv.Itoa(chatID))
	qr.Add("photo", photoURL)
	qr.Add("caption", caption)

	_, err := c.doRequest(sendPhotoMethod, qr)
	if err != nil {
		return er.Wrap("sending photo failed", err)
	}
	return nil
}

func (c *Client) SendVideo(chatID int, videoURL string, caption string) error {
	qr := url.Values{}
	qr.Add("chat_id", strconv.Itoa(chatID))
	qr.Add("video", videoURL)
	qr.Add("caption", caption)

	_, err := c.doRequest(sendVideoMethod, qr)
	if err != nil {
		return er.Wrap("sending video failed", err)
	}
	return nil
}

func (c *Client) SendMediaGroup(chatID int, mediaGr []MediaGroup) error {
	res, err := json.Marshal(mediaGr)
	if err != nil {
		return er.Wrap("media group marshalling failed", err)
	}
	mediaArray := string(res)

	qr := url.Values{}
	qr.Add("chat_id", strconv.Itoa(chatID))
	qr.Add("media", mediaArray)

	_, err = c.doRequest(sendMediaGroupMethod, qr)
	if err != nil {
		return er.Wrap("sending carousel failed", err)
	}
	return nil
}

func (c *Client) SendPost(chatIDid int, mType []int, urls []string, caption string) error {
	mediaGr := NewMediaGr(mType, urls, caption)
	if len(mType) > 1 {
		return c.SendMediaGroup(chatIDid, *mediaGr)
	}

	switch mType[0] {
	case 1:
		return c.SendPhoto(chatIDid, urls[0], caption)
	case 2:
		return c.SendVideo(chatIDid, urls[0], caption)
	default:
		return errors.New("unknown media type")
	}
}
func NewMediaGr(mType []int, urls []string, caption string) *[]MediaGroup {
	mediaGR := make([]MediaGroup, 0, len(mType))
	for i := 0; i < len(mType); i++ {
		mg := MediaGroup{
			ContentType: mTypeValue(mType[i]),
			ContentURL:  urls[i],
			Caption:     captionValue(i, caption),
		}
		mediaGR = append(mediaGR, mg)
	}
	return &mediaGR
}

func mTypeValue(mType int) string {
	switch mType {
	case 1:
		return "photo"
	case 2:
		return "video"
	default:
		return ""
	}
}

func captionValue(i int, cap string) interface{} {
	if i == 0 {
		return cap
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

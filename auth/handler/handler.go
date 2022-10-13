package handler

import (
	"encoding/json"
	"github.com/EgorMamoshkin/InstaBot/auth"
	"github.com/EgorMamoshkin/InstaBot/clients/tgclient"
	"github.com/EgorMamoshkin/InstaBot/events/telegram"
	"github.com/EgorMamoshkin/InstaBot/lib/er"
	"github.com/EgorMamoshkin/InstaBot/storage"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
)

type ResponseHandler struct {
	host     string
	client   http.Client
	tgClient *tgclient.Client
	storage  storage.Storage
}

func New(host string, tgClient *tgclient.Client, storage storage.Storage) *ResponseHandler {
	return &ResponseHandler{
		host:     host,
		client:   http.Client{},
		tgClient: tgClient,
		storage:  storage,
	}
}

func (rh *ResponseHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	chatIDStr := r.URL.Query().Get("chat_id")
	code := strings.TrimRight(r.URL.Query().Get("code"), "#_")

	qr := url.Values{}
	qr.Add("client_id", "")     // TODO : Change getting from flag
	qr.Add("client_secret", "") // TODO : Change getting from flag
	qr.Add("grant_type", "authorization_code")
	qr.Add("redirect_uri", "localhost/auth") // TODO : Change
	qr.Add("code", code)

	data, err := rh.doRequest(qr)
	if err != nil {
		log.Printf("can't get access token: %s", err)
	}

	var userToken auth.UserAccess

	err = json.Unmarshal(data, &userToken)
	if err != nil {
		log.Printf("can't unmarshal response: %s", err)
	}

	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		log.Printf("can't convert chatID to type int: %s", err)
	}

	err = rh.storage.SaveToken(chatID, userToken)
	if err != nil {
		log.Printf("can't save token: %s", err)
	} else {
		err = rh.tgClient.SendMessage(chatID, telegram.MsgSuccessfulAuth)
		if err != nil {
			log.Printf("can't send message: %s", err)
		}
	}
}

func (rh *ResponseHandler) doRequest(query url.Values) ([]byte, error) {
	const reqError = "request failed"

	u := url.URL{
		Scheme: "https",
		Host:   rh.host,
		Path:   path.Join("oauth", "access_token"),
	}

	req, err := http.NewRequest(http.MethodPost, u.String(), nil)
	if err != nil {
		return nil, er.Wrap(reqError, err)
	}

	req.URL.RawQuery = query.Encode()

	resp, err := rh.client.Do(req)
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

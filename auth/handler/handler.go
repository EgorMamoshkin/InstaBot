package handler

import (
	"encoding/json"
	"errors"
	"fmt"
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
	chatID, err := rh.handle(r)
	if err != nil {
		log.Printf("can't handle request: %s", err)
		err = rh.tgClient.SendMessage(chatID, telegram.MsgAuthFailed)
		if err != nil {
			log.Printf("can't send message: %s", err)
		}
	} else {
		err = rh.tgClient.SendMessage(chatID, telegram.MsgSuccessfulAuth)
		if err != nil {
			log.Printf("can't send message: %s", err)
		}
	}
}

func (rh *ResponseHandler) TokenRedirect(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimRight(r.URL.Query().Get("code"), "#_")
	if code == "" {
		log.Println("there are no code")
	}

	link := fmt.Sprintf("https://telegram.me/share/url?url=/requesttoken&text=%s", code)

	http.Redirect(w, r, link, http.StatusSeeOther)
}

func (rh *ResponseHandler) handle(r *http.Request) (int, error) {
	chatIDStr := r.URL.Query().Get("chat_id")
	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		return chatID, er.Wrap("can't convert chatID to type int: %s", err)
	}

	code := strings.TrimRight(r.URL.Query().Get("code"), "#_")
	if code == "" {
		return chatID, errors.New("there are no code")
	}

	qr := url.Values{}
	qr.Add("client_id", telegram.AppID) // TODO : Change getting from flag
	qr.Add("client_secret", "")         // TODO : Change getting from flag
	qr.Add("grant_type", "authorization_code")
	qr.Add("redirect_uri", "https://188.225.60.154:8080/auth") // TODO : Change
	qr.Add("code", code)

	data, err := rh.doRequest(qr)
	if err != nil {
		return chatID, er.Wrap("can't get access token: %s", err)
	}

	var userToken auth.UserAccess

	err = json.Unmarshal(data, &userToken)
	if err != nil {
		return chatID, er.Wrap("can't unmarshal response: %s", err)
	}

	if userToken.UserID == 0 || userToken.Token == "" {
		return chatID, errors.New(fmt.Sprintf("instagram refused to receive a token: %s", string(data)))
	}

	err = rh.storage.SaveToken(chatID, userToken)
	if err != nil {
		return chatID, er.Wrap("can't save token: ", err)
	}

	return chatID, nil
}

func (rh *ResponseHandler) doRequest(query url.Values) ([]byte, error) {
	const reqError = "request failed"

	u := url.URL{
		Scheme: "https",
		Host:   rh.host,
		Path:   path.Join("oauth", "access_token"),
	}
	/*
		req, err := http.NewRequest(http.MethodPost, u.String(), nil)
		if err != nil {
			return nil, er.Wrap(reqError, err)
		}

		req.URL.RawQuery = query.Encode()

		fmt.Println(req)

		resp, err := rh.client.Do(req)
		if err != nil {
			return nil, er.Wrap(reqError, err)
		}

	*/

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

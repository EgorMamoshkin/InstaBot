package telegram

import (
	"context"
	"errors"
	"fmt"
	"github.com/Davincible/goinsta/v3"
	"github.com/EgorMamoshkin/InstaBot/apiclient/instagramapi"
	"github.com/EgorMamoshkin/InstaBot/clients/tgclient"
	insta_parse "github.com/EgorMamoshkin/InstaBot/insta-parse"
	"github.com/EgorMamoshkin/InstaBot/lib/er"
	"github.com/EgorMamoshkin/InstaBot/storage"
	"log"
	"net/url"
	"path"
	"strings"
	"sync"
)

const (
	HelpCmd        = "/help"
	StartCmd       = "/start"
	GetUpdatesCmd  = "/upd"
	StartAuth      = "/startauth"
	GetAccessToken = "/getaccess"
	GetPosts       = "/getposts"
)

func (p *Processor) execCmd(ctx context.Context, text string, chatID int, username string) error {
	text = strings.TrimSpace(text)

	log.Printf("new command '%s' from %s(%d)", text, username, chatID)

	command, ok := isCommand(text)
	if ok {
		text = command[0]
	}

	switch text {
	case HelpCmd:
		return p.SendHelp(chatID)
	case StartCmd:
		return p.SendHello(chatID)
	case GetUpdatesCmd:
		return p.StartFeedUpd(ctx, chatID, username)
	case StartAuth:
		return p.StartAuth(chatID)
	case GetAccessToken:
		return p.AccessToken(ctx, chatID, command[1])
	case GetPosts:
		return p.GetPosts(ctx, chatID)
	default:
		if login, pass, err := isLoginPass(text); err != nil {
			_ = p.tg.SendMessage(chatID, msgUnknownCommand)

			return err
		} else {
			return p.SaveInstAcc(ctx, chatID, login, pass, username)
		}
	}
}

func (p *Processor) SendHelp(chatID int) error {
	return p.tg.SendMessage(chatID, msgHelp)
}

func (p *Processor) SendHello(chatID int) error {
	return p.tg.SendMessage(chatID, msgHello)
}

func (p *Processor) SaveInstAcc(ctx context.Context, chatID int, login string, pass string, username string) error {
	instAcc, err := loginInstagram(login, pass)
	if err != nil {
		_ = p.tg.SendMessage(chatID, msgLogInFailed)

		return er.Wrap("log in to account failed:", err)
	}

	lastID := lastPostID(instAcc)

	user := storage.User{
		LastPostID: lastID,
		InstAcc:    instAcc,
	}

	if err := p.storage.SaveAccount(ctx, &user, username); err != nil {
		_ = p.tg.SendMessage(chatID, msgSavingAccFailed)

		return er.Wrap("account saving failed: ", err)
	}

	return p.tg.SendMessage(chatID, msgLoggedIn)
}

func (p *Processor) StartFeedUpd(ctx context.Context, chatID int, username string) error {
	user, err := p.storage.GetAccount(ctx, username)
	if err != nil {
		_ = p.tg.SendMessage(chatID, msgOpenAccFailed)

		return er.Wrap("can't get your account", err)
	}

	if user == nil {
		return p.tg.SendMessage(chatID, msgNotLoggedInBefore)
	}

	log.Printf("user %s logged in\n", username)

	timeLine := user.InstAcc.Timeline
	lastPID := user.LastPostID

	if ok, err := timeLine.NewFeedPostsExist(); ok {
		log.Println("New post available")

		if err != nil {
			log.Println(err)
		}

		np := newPosts(timeLine, lastPID)
		for _, post := range np {
			if post.IsSeen {
				continue
			}

			mType, urls, caption, err := insta_parse.GetData(post)
			if err != nil {
				log.Println(err)

				continue
			}
			_ = p.tg.SendPost(chatID, mType, urls, caption)
		}

		if len(np) != 0 {
			return p.storage.SaveLastPostID(ctx, np[0].ID.(string), username)
		}

		return p.tg.SendMessage(chatID, msgNoNewPost)
	}

	return p.tg.SendMessage(chatID, msgNoNewPost)
}

func (p *Processor) StartAuth(chatID int) error {
	requestURL := url.URL{
		Scheme: "https",
		Host:   p.inst.GetAPIHost(),
		Path:   path.Join("oauth", "authorize"),
	}

	query := url.Values{}
	query.Add("client_id", p.inst.GetAppID())
	query.Add("redirect_uri", p.inst.GetRedirectURI())
	query.Add("scope", "user_profile,user_media")
	query.Add("response_type", "code")

	requestURL.RawQuery = query.Encode()

	return p.tg.SendMessage(chatID, requestURL.String())
}

func (p *Processor) AccessToken(ctx context.Context, chatID int, reqToken string) error {
	userToken, err := p.inst.ExchangeToAccessToken(reqToken)
	if err != nil {
		_ = p.tg.SendMessage(chatID, msgAuthFailed)

		return er.Wrap("can't get access token: ", err)
	}

	ok, err := p.storage.IsUserExist(ctx, chatID)
	if err != nil {
		log.Println("can't check is user exists: ", err)
	}

	if ok {
		err = p.storage.UpdateToken(ctx, chatID, userToken)
		if err != nil {
			return er.Wrap("can't update token: ", err)
		}
	} else {
		err = p.storage.SaveToken(ctx, chatID, userToken)
		if err != nil {
			_ = p.tg.SendMessage(chatID, msgAuthFailed)

			return er.Wrap("can't save token: ", err)
		}
	}

	return p.tg.SendMessage(chatID, msgSuccessfulAuth)
}

func (p *Processor) GetPosts(ctx context.Context, chatID int) error {
	instUser, err := p.storage.GetInstUser(ctx, chatID)
	if err != nil {
		_ = p.tg.SendMessage(chatID, msgCantGetPosts)

		return er.Wrap("can't get user from DB: ", err)
	}

	userPosts, err := p.inst.GetPosts(instUser)
	if err != nil {
		return er.Wrap("can't get user posts:", err)
	}

	var wg sync.WaitGroup

	wg.Add(len(userPosts.Data))

	for _, post := range userPosts.Data {
		go func(post instagramapi.Media) {
			caption := fmt.Sprintf("@%s\n%s", post.Username, post.Caption)

			switch post.MediaType {
			case "IMAGE":
				_ = p.tg.SendPhoto(chatID, post.MediaURL, caption)
			case "VIDEO":
				_ = p.tg.SendVideo(chatID, post.MediaURL, caption)
			case "CAROUSEL_ALBUM":
				carousel, err := p.inst.CarouselElements(post.ID, instUser)
				if err != nil {
					log.Println(er.Wrap("can't get carousel data: ", err))
				}
				_ = p.tg.SendMediaGroup(chatID, tgclient.CreateMediaGr(carousel, caption))
				_ = p.tg.SendMessage(chatID, caption)
			default:
				_ = p.tg.SendMessage(chatID, fmt.Sprintf("%s type of post doesn't supply", post.MediaType))
			}

			wg.Done()
		}(post)
	}

	wg.Wait()

	return nil
}

func loginInstagram(login string, pass string) (*goinsta.Instagram, error) {
	instAcc := goinsta.New(login, pass)
	if err := instAcc.Login(); err != nil {
		return nil, err
	}

	return instAcc, nil
}

func isLoginPass(text string) (string, string, error) {
	errWrongFormat := errors.New("incorrect login and password input format")

	if !strings.HasPrefix(text, "LOG:") || !strings.Contains(text, "PASS:") {
		return "", "", errWrongFormat
	}

	text = strings.TrimLeft(text, "LOG:")

	passLog := strings.Split(text, "PASS:")
	if len(passLog) != 2 {
		return "", "", errWrongFormat
	}

	return passLog[0], passLog[1], nil
}

func lastPostID(instAcc *goinsta.Instagram) string {
	tLine := instAcc.Timeline
	tLine.Next()
	posts := tLine.Items

	return posts[0].ID.(string)
}

func newPosts(tLine *goinsta.Timeline, lastPID string) []*goinsta.Item {
	posts := tLine.Items
	idx := -1

	for idx == -1 {
		idx = getIndex(lastPID, posts)
		if idx == -1 {
			tLine.Next()
			posts = tLine.Items
		}
	}

	return posts[:idx]
}

func getIndex(lastItemID string, posts []*goinsta.Item) int {
	for idx, post := range posts {
		if lastItemID == post.ID {
			return idx
		}
	}

	return -1
}

func isCommand(text string) ([]string, bool) {
	if strings.HasPrefix(text, "/") {
		command := strings.Split(text, "\n")

		return command, true
	}

	return nil, false
}

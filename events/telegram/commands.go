package telegram

import (
	insta_parse "InstaBot/insta-parse"
	"InstaBot/lib/er"
	"InstaBot/storage"
	"context"
	"errors"
	"github.com/Davincible/goinsta/v3"
	"log"
	"strings"
)

const (
	HelpCmd       = "/help"
	StartCmd      = "/start"
	GetUpdatesCmd = "/upd"
)

func (p *Processor) execCmd(ctx context.Context, text string, chatID int, username string) error {
	text = strings.TrimSpace(text)

	log.Printf("new command '%s' from %s(%d)", text, username, chatID)

	switch text {
	case HelpCmd:
		return p.SendHelp(chatID)
	case StartCmd:
		return p.SendHello(chatID)
	case GetUpdatesCmd:
		return p.StartFeedUpd(ctx, chatID, username)
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

package storage

import (
	"context"
	"github.com/Davincible/goinsta/v3"
	"github.com/EgorMamoshkin/InstaBot/apiclient/instagramapi"
)

type Storage interface {
	SaveAccount(ctx context.Context, u *User, lastPID string) error
	GetAccount(ctx context.Context, userName string) (*User, error)
	SaveLastPostID(ctx context.Context, postID string, username string) error
	SaveToken(ctx context.Context, chatID int, userToken *instagramapi.User) error
	IsUserExist(ctx context.Context, chatID int) (bool, error)
	UpdateToken(ctx context.Context, chatID int, user *instagramapi.User) error
}

type User struct {
	LastPostID string
	InstAcc    *goinsta.Instagram
}

package storage

import (
	"github.com/Davincible/goinsta/v3"
)

type Storage interface {
	SaveAccount(u *User, lastPID string) error
	GetAccount(userName string) (*User, error)
	SaveLastPostID(postID string, username string) error
}

type User struct {
	LastPostID string
	InstAcc    *goinsta.Instagram
}

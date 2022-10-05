package storage

import (
	"github.com/Davincible/goinsta/v3"
)

type Storage interface {
	SaveAccount(u *User) error
	GetAccount(userName string) (*User, error)
}

type User struct {
	UserName string
	InstAcc  *goinsta.Instagram
}

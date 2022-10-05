package files

import (
	"InstaBot/lib/er"
	"InstaBot/storage"
	"fmt"
	"github.com/Davincible/goinsta/v3"
	"os"
	"path/filepath"
)

type Storage struct {
	basePath string
}

const defaultPerm = 0744

func New(basePath string) Storage {
	return Storage{basePath: basePath}
}

func (s Storage) SaveAccount(user *storage.User) error {
	fPath := filepath.Join(s.basePath, user.UserName)

	if err := os.MkdirAll(fPath, defaultPerm); err != nil {
		return er.Wrap("creating directories failed:", err)
	}

	fPath = filepath.Join(fPath, fileName(user.UserName))
	err := user.InstAcc.Export(fPath)
	if err != nil {
		return er.Wrap("account saving failed:", err)
	}
	return nil
}

func (s Storage) GetAccount(userName string) (*storage.User, error) {
	fPath := filepath.Join(s.basePath, userName, fileName(userName))

	instAcc, err := goinsta.Import(fPath)
	if err != nil {
		return nil, er.Wrap("instagram account import failed:", err)
	}

	return &storage.User{UserName: userName, InstAcc: instAcc}, nil
}

func fileName(userName string) string {
	return fmt.Sprintf("%s.json", userName)
}

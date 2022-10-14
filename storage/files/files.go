package files

import (
	"fmt"
	"github.com/Davincible/goinsta/v3"
	"github.com/EgorMamoshkin/InstaBot/lib/er"
	"github.com/EgorMamoshkin/InstaBot/storage"
	"io/ioutil"
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

func (s Storage) SaveAccount(user *storage.User, username string) error {
	instAcFPath := filepath.Join(s.basePath, username)

	if err := os.MkdirAll(instAcFPath, defaultPerm); err != nil {
		return er.Wrap("creating directories failed:", err)
	}

	instAcFPath = filepath.Join(instAcFPath, instFileName(username))

	err := user.InstAcc.Export(instAcFPath)
	if err != nil {
		return er.Wrap("account saving failed:", err)
	}

	lastPostFPath := filepath.Join(instAcFPath, lastPostFileName(username))

	err = saveLastPostID(lastPostFPath, user.LastPostID)
	if err != nil {
		return er.Wrap("saving last post ID failed:", err)
	}

	return nil
}

func (s Storage) GetAccount(username string) (*storage.User, error) {
	fPath := filepath.Join(s.basePath, username, instFileName(username))

	instAcc, err := goinsta.Import(fPath)
	if err != nil {
		return nil, er.Wrap("instagram account import failed:", err)
	}

	lastPostFPath := filepath.Join(s.basePath, username, lastPostFileName(username))

	lastID, err := readLastPostID(lastPostFPath)
	if err != nil {
		return nil, er.Wrap("getting last post ID failed:", err)
	}

	return &storage.User{LastPostID: lastID, InstAcc: instAcc}, nil
}

func (s Storage) SaveLastPostID(postID string, username string) error {
	lastPostFPath := filepath.Join(s.basePath, username, lastPostFileName(username))

	err := saveLastPostID(lastPostFPath, postID)
	if err != nil {
		return er.Wrap("saving last post ID failed:", err)
	}

	return nil
}

func instFileName(userName string) string {
	return fmt.Sprintf("%s.json", userName)
}

func lastPostFileName(userName string) string {
	return fmt.Sprintf("ID%s.txt", userName)
}

func saveLastPostID(fPath string, lastID string) error {
	file, err := os.Create(fPath)
	if err != nil {
		return er.Wrap("creating file failed: ", err)
	}

	_, err = file.WriteString(lastID)
	if err != nil {
		return er.Wrap("writing to file failed: ", err)
	}

	err = file.Close()
	if err != nil {
		return er.Wrap("closing file failed: ", err)
	}

	return nil
}

func readLastPostID(fPath string) (string, error) {
	id, err := ioutil.ReadFile(fPath)
	if err != nil {
		return "", er.Wrap("can't open file: ", err)
	}

	return string(id), nil
}

package files

import (
	"InstaBot/lib/er"
	"InstaBot/storage"
	"context"
	"database/sql"
	"encoding/json"
	"github.com/Davincible/goinsta/v3"
	_ "github.com/jackc/pgx/stdlib"
)

type Storage struct {
	db *sql.DB
}

func New(dsn string) (*Storage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, er.Wrap("can't open database: ", err)
	}
	if err := db.Ping(); err != nil {
		return nil, er.Wrap("can't connect to database: ", err)
	}
	return &Storage{db: db}, nil
}

func (s *Storage) SaveAccount(ctx context.Context, user *storage.User, username string) error {

	config := user.InstAcc.ExportConfig()
	configJson, err := json.Marshal(config)
	instConfig := string(configJson)
	if err != nil {
		return er.Wrap("can't convert config into json:", err)
	}

	q := `INSERT INTO instagram_users (username_tg, instagram_acc, last_post_id) VALUES ($1, $2, $3)`
	_, err = s.db.ExecContext(ctx, q, username, instConfig, user.LastPostID)
	if err != nil {
		return er.Wrap("can't save account data into database: ", err)
	}
	return nil
}

func (s *Storage) GetAccount(ctx context.Context, username string) (*storage.User, error) {

	q := `SELECT instagram_acc, last_post_id FROM instagram_users WHERE username_tg = $1`

	row := s.db.QueryRowContext(ctx, q, username)

	var instConfig string
	var lastPostID string

	err := row.Scan(&instConfig, &lastPostID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, er.Wrap("can't get instagram account: ", err)
	}

	config := goinsta.ConfigFile{}

	err = json.Unmarshal([]byte(instConfig), &config)
	if err != nil {
		return nil, er.Wrap("can't convert json to ConfigFile: ", err)
	}

	instAcc, err := goinsta.ImportConfig(config)
	if err != nil {
		return nil, er.Wrap("can't import instagram account: ", err)
	}

	return &storage.User{LastPostID: lastPostID, InstAcc: instAcc}, nil
}

func (s *Storage) SaveLastPostID(ctx context.Context, postID string, username string) error {
	q := `UPDATE instagram_users SET last_post_id = $1 WHERE username_tg = $2`

	_, err := s.db.ExecContext(ctx, q, postID, username)
	if err == sql.ErrNoRows {
		return er.Wrap("can't save last post ID: ", err)
	}
	return nil
}

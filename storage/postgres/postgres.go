package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/Davincible/goinsta/v3"
	"github.com/EgorMamoshkin/InstaBot/apiclient/instagramapi"
	"github.com/EgorMamoshkin/InstaBot/lib/er"
	"github.com/EgorMamoshkin/InstaBot/storage"
	_ "github.com/jackc/pgx/v5/stdlib"
	"time"
)

type Storage struct {
	db *sql.DB
}

// New creates new PostgresSQL storage.
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

// SaveAccount Saves account in storage.
func (s *Storage) SaveAccount(ctx context.Context, user *storage.User, username string) error {
	config := user.InstAcc.ExportConfig()

	configJSON, err := json.Marshal(config)
	if err != nil {
		return er.Wrap("can't convert config into json:", err)
	}

	instConfig := string(configJSON)

	ctx, _ = context.WithTimeout(ctx, 5*time.Second)

	q := `INSERT INTO instagram_users (username_tg, instagram_acc, last_post_id) VALUES ($1, $2, $3)`

	_, err = s.db.ExecContext(ctx, q, username, instConfig, user.LastPostID)
	if err != nil {
		return er.Wrap("can't save account data into database: ", err)
	}

	return nil
}

// GetAccount imports account from storage.
func (s *Storage) GetAccount(ctx context.Context, username string) (*storage.User, error) {
	ctx, _ = context.WithTimeout(ctx, 5*time.Second)

	q := `SELECT instagram_acc, last_post_id FROM instagram_users WHERE username_tg = $1`
	row := s.db.QueryRowContext(ctx, q, username)

	var instConfig string

	var lastPostID string

	err := row.Scan(&instConfig, &lastPostID)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, er.Wrap("can't get instagramapi account: ", err)
	}

	config := goinsta.ConfigFile{}

	err = json.Unmarshal([]byte(instConfig), &config)
	if err != nil {
		return nil, er.Wrap("can't convert json to ConfigFile: ", err)
	}

	instAcc, err := goinsta.ImportConfig(config)
	if err != nil {
		return nil, er.Wrap("can't import instagramapi account: ", err)
	}

	return &storage.User{LastPostID: lastPostID, InstAcc: instAcc}, nil
}

// SaveLastPostID saves last post from feed in storage.
func (s *Storage) SaveLastPostID(ctx context.Context, postID string, username string) error {
	ctx, _ = context.WithTimeout(ctx, 5*time.Second)

	q := `UPDATE instagram_users SET last_post_id = $1 WHERE username_tg = $2`

	_, err := s.db.ExecContext(ctx, q, postID, username)
	if err != nil {
		return er.Wrap("can't save last post ID: ", err)
	}

	return nil
}

func (s *Storage) SaveToken(ctx context.Context, chatID int, userToken *instagramapi.User) error {
	q := `INSERT INTO token(tg_chat_id, instagram_user_id, access_token) VALUES($1, $2, $3)`

	ctx, _ = context.WithTimeout(ctx, 5*time.Second)

	_, err := s.db.ExecContext(ctx, q, chatID, userToken.UserID, userToken.Token)
	if err != nil {
		return er.Wrap("can't save token: ", err)
	}

	return nil
}

func (s *Storage) Init(ctx context.Context) error {
	q := `CREATE TABLE IF NOT EXISTS instagram_users (user_id SERIAL PRIMARY KEY, username_tg VARCHAR(40), instagram_acc VARCHAR, last_post_id VARCHAR(40));
	CREATE TABLE IF NOT EXISTS token (id SERIAL PRIMARY KEY, tg_chat_id BIGINT, instagram_user_id BIGINT, access_token VARCHAR)`

	ctx, _ = context.WithTimeout(ctx, 5*time.Second)

	_, err := s.db.ExecContext(ctx, q)
	if err != nil {
		return er.Wrap("can't create new table: ", err)
	}

	return nil
}

func (s *Storage) IsUserExist(ctx context.Context, chatID int) (bool, error) {
	q := `SELECT COUNT(tg_chat_id) FROM token WHERE tg_chat_id = $1`

	ctx, _ = context.WithTimeout(ctx, 5*time.Second)

	row := s.db.QueryRowContext(ctx, q, chatID)

	var res int

	err := row.Scan(&res)
	if err != nil {
		return false, er.Wrap("scaning rows ERROR: ", err)
	}

	return res > 0, nil
}

func (s *Storage) UpdateToken(ctx context.Context, chatID int, user *instagramapi.User) error {
	q := `UPDATE token SET instagram_user_id = $1, access_token = $2 WHERE tg_chat_id = $3`

	ctx, _ = context.WithTimeout(ctx, 5*time.Second)

	_, err := s.db.ExecContext(ctx, q, user.UserID, user.Token, chatID)
	if err != nil {
		return er.Wrap("can't update token: ", err)
	}

	return nil
}

func (s *Storage) GetInstUser(ctx context.Context, chatID int) (*instagramapi.User, error) {
	q := `SELECT instagram_user_id, access_token FROM token WHERE tg_chat_id = $1`

	row := s.db.QueryRowContext(ctx, q, chatID)

	var user instagramapi.User

	err := row.Scan(&user.UserID, &user.Token)
	if err == sql.ErrNoRows {
		return nil, er.Wrap("there are no any saved accounts by this user: ", err)
	}

	if err != nil {
		return nil, er.Wrap("can't get user's account: ", err)
	}

	return &user, nil
}

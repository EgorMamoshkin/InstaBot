package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/Davincible/goinsta/v3"
	"github.com/EgorMamoshkin/InstaBot/auth"
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

// SaveLastPostID saves last post from feed in storage.
func (s *Storage) SaveLastPostID(ctx context.Context, postID string, username string) error {
	ctx, _ = context.WithTimeout(ctx, 5*time.Second)

	q := `UPDATE instagram_users SET last_post_id = $1 WHERE username_tg = $2`

	_, err := s.db.ExecContext(ctx, q, postID, username)
	if err == sql.ErrNoRows {
		return er.Wrap("can't save last post ID: ", err)
	}

	return nil
}

func (s *Storage) SaveToken(chatID int, userToken auth.UserAccess) error {
	q := `INSERT INTO token(tg_chat_id, instagram_user_id, access_token) VALUES($1, $2, $3)`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

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

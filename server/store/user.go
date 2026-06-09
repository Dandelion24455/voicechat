package store

import (
	"context"
	"voicechat-server/model"

	"github.com/jackc/pgx/v5"
)

func (db *DB) CreateUser(ctx context.Context, u *model.User) error {
	_, err := db.Pool.Exec(ctx,
		"INSERT INTO users (id, player_id, username, password_hash) VALUES ($1, $2, $3, $4)",
		u.ID, u.PlayerID, u.Username, u.PasswordHash,
	)
	return err
}

func (db *DB) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	u := &model.User{}
	err := db.Pool.QueryRow(ctx,
		"SELECT id, COALESCE(player_id,''), username, password_hash, created_at FROM users WHERE username=$1",
		username,
	).Scan(&u.ID, &u.PlayerID, &u.Username, &u.PasswordHash, &u.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func (db *DB) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	u := &model.User{}
	err := db.Pool.QueryRow(ctx,
		"SELECT id, COALESCE(player_id,''), username, password_hash, created_at FROM users WHERE id=$1",
		id,
	).Scan(&u.ID, &u.PlayerID, &u.Username, &u.PasswordHash, &u.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func (db *DB) UpdatePlayerID(ctx context.Context, userID, playerID string) error {
	_, err := db.Pool.Exec(ctx,
		"UPDATE users SET player_id=$1 WHERE id=$2 AND player_id IS NULL",
		playerID, userID,
	)
	return err
}

func (db *DB) IsPlayerIDTaken(ctx context.Context, playerID string) (bool, error) {
	var exists bool
	err := db.Pool.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM users WHERE UPPER(player_id)=$1)", playerID,
	).Scan(&exists)
	return exists, err
}

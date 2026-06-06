package store

import (
	"context"
	"voicechat-server/model"

	"github.com/jackc/pgx/v5"
)

func (db *DB) CreateRoom(ctx context.Context, r *model.Room) error {
	_, err := db.Pool.Exec(ctx,
		"INSERT INTO rooms (id, name, creator_id, invite_code) VALUES ($1, $2, $3, $4)",
		r.ID, r.Name, r.CreatorID, r.InviteCode,
	)
	return err
}

func (db *DB) ListRooms(ctx context.Context) ([]model.Room, error) {
	rows, err := db.Pool.Query(ctx,
		"SELECT id, name, creator_id, COALESCE(invite_code,''), created_at FROM rooms ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []model.Room
	for rows.Next() {
		var r model.Room
		if err := rows.Scan(&r.ID, &r.Name, &r.CreatorID, &r.InviteCode, &r.CreatedAt); err != nil {
			return nil, err
		}
		rooms = append(rooms, r)
	}
	return rooms, nil
}

func (db *DB) ListRoomsByUser(ctx context.Context, userID string) ([]model.Room, error) {
	rows, err := db.Pool.Query(ctx,
		"SELECT id, name, creator_id, COALESCE(invite_code,''), created_at FROM rooms WHERE creator_id=$1 ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []model.Room
	for rows.Next() {
		var r model.Room
		if err := rows.Scan(&r.ID, &r.Name, &r.CreatorID, &r.InviteCode, &r.CreatedAt); err != nil {
			return nil, err
		}
		rooms = append(rooms, r)
	}
	return rooms, nil
}

func (db *DB) GetRoomByID(ctx context.Context, id string) (*model.Room, error) {
	r := &model.Room{}
	err := db.Pool.QueryRow(ctx,
		"SELECT id, name, creator_id, COALESCE(invite_code,''), created_at FROM rooms WHERE id=$1", id,
	).Scan(&r.ID, &r.Name, &r.CreatorID, &r.InviteCode, &r.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return r, err
}

func (db *DB) GetRoomByInviteCode(ctx context.Context, code string) (*model.Room, error) {
	r := &model.Room{}
	err := db.Pool.QueryRow(ctx,
		"SELECT id, name, creator_id, COALESCE(invite_code,''), created_at FROM rooms WHERE UPPER(invite_code)=$1", code,
	).Scan(&r.ID, &r.Name, &r.CreatorID, &r.InviteCode, &r.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return r, err
}

func (db *DB) DeleteRoom(ctx context.Context, id string, userID string) error {
	tag, err := db.Pool.Exec(ctx,
		"DELETE FROM rooms WHERE id=$1 AND creator_id=$2", id, userID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

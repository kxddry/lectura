package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/kxddry/lectura/shared/entities/config/db"
	"github.com/kxddry/lectura/shared/entities/frontend"
	"github.com/kxddry/lectura/shared/entities/summarized"
	"github.com/kxddry/lectura/shared/entities/transcribed"
	"github.com/kxddry/lectura/shared/entities/uploaded"
	"github.com/kxddry/lectura/shared/utils/storage"
	"github.com/lib/pq"
)

type Client struct {
	db *sql.DB
}

func New(cfg db.StorageConfig) (*Client, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)
	_db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return &Client{db: _db}, _db.Ping()
}

func (c *Client) Close() error { return c.db.Close() }

func (c *Client) AddFile(ctx context.Context, msg uploaded.Record) error {
	const op = "storage.postgres.addFile"

	tx, err := c.db.Begin()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback()

	row := tx.QueryRowContext(ctx, `INSERT INTO files (uuid, user_id, og_filename, og_extension, status) VALUES ($1, $2, $3, $4, $5);`,
		msg.UUID, msg.Update.UserID, msg.Update.OGFileName, msg.Update.OGExtension, 0,
	)
	if err = row.Err(); err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "unique_violation" {
			return fmt.Errorf("%s: %w", op, storage.ErrUUIDExists)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return tx.Commit()
}

func (c *Client) AddTranscription(ctx context.Context, msg transcribed.Record) error {
	const op = "storage.postgres.addTranscription"
	tx, err := c.db.Begin()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(ctx, `INSERT INTO transcribed (uuid, text, language) VALUES ($1, $2, $3)`, msg.UUID, msg.Text, msg.Language).Err()
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "unique_violation" {
			return fmt.Errorf("%s: %w", op, storage.ErrUUIDExists)
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return tx.Commit()
}

func (c *Client) AddSummarization(ctx context.Context, msg summarized.Record) error {
	const op = "storage.postgres.addSummarization"
	tx, err := c.db.Begin()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(ctx, `INSERT INTO summarized (uuid, text) VALUES ($1, $2)`, msg.UUID, msg.Text).Err()
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "unique_violation" {
			return fmt.Errorf("%s: %w", op, storage.ErrUUIDExists)
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return tx.Commit()
}

func (c *Client) UpdateFile(ctx context.Context, uuid string, status int) error {
	const op = "storage.postgres.updateFile"
	tx, err := c.db.Begin()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback()

	var id int
	err = tx.QueryRowContext(ctx, `SELECT status FROM files WHERE uuid = $1`, uuid).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.ErrUUIDNotFound
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	if id > status {
		return fmt.Errorf("%s: %w", storage.ErrNewerStatus)
	}

	row := tx.QueryRowContext(ctx, `UPDATE files SET status = $1 WHERE uuid = $2`, status, uuid)
	if err = row.Err(); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%s: %w", storage.ErrUUIDNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return tx.Commit()
}

func (c *Client) DeleteFile(ctx context.Context, uuid string) error {
	const op = "storage.postgres.deleteFile"
	tx, err := c.db.Begin()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback()
	row := tx.QueryRowContext(ctx, `DELETE FROM files WHERE uuid = $1`, uuid)
	if err = row.Err(); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%s: %w", storage.ErrUUIDNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return tx.Commit()
}

func (c *Client) ListFiles(ctx context.Context, user_id uint) ([]frontend.File, error) {
	const op = "storage.postgres.listFiles"
	tx, err := c.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `SELECT og_filename, og_extension, uuid, status FROM files WHERE user_id = $1;`, user_id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()
	var files []frontend.File
	for rows.Next() {
		var uuid, ext, name string
		var status uint8
		if err := rows.Scan(&name, &ext, &uuid, &status); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		files = append(files, frontend.File{
			UUID:     uuid,
			Name:     name + ext,
			URL:      "",
			MimeType: uploaded.Extensions[ext],
			Status:   status,
		})
	}
	if len(files) == 0 {
		return nil, storage.ErrNoFiles
	}
	return files, nil
}

func (c *Client) GetFileData(ctx context.Context, uuid string, uid uint) (string, error) {
	const op = "storage.postgres.getFileData"
	tx, err := c.db.Begin()
	if err != nil {
		return err.Error(), fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback()

	var status int
	err = tx.QueryRowContext(ctx, `SELECT status FROM files WHERE uuid = $1 AND user_id = $2;`, uuid, uid).Scan(&status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "This file does not exist", nil
		}
		return err.Error(), fmt.Errorf("%s: %w", op, err)
	}

	if status == 0 {
		return "Your file has not been processed yet, please wait...", nil
	} else if status == 1 {
		var data string
		err = tx.QueryRowContext(ctx, `SELECT text FROM transcribed WHERE uuid = $1;`, uuid).Scan(&data)
		if err != nil {
			return err.Error(), fmt.Errorf("%s: %w", op, err)
		}
		return data, nil
	} else {
		var data string
		err = tx.QueryRowContext(ctx, `SELECT text FROM summarized WHERE uuid = $1;`, uuid).Scan(&data)
		if err != nil {
			return err.Error(), fmt.Errorf("%s: %w", op, err)
		}
		return data, nil
	}
}

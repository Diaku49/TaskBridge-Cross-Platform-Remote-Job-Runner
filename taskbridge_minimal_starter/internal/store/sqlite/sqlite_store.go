package sqlite

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"taskbridge/internal/store/sqlite/generated"

	_ "modernc.org/sqlite"
)

type SqliteStore struct {
	db *sql.DB
	q  *generated.Queries
}

//go:embed schema.sql
var schemaSQL string

func NewSqliteStore(ctx context.Context, path string) (*SqliteStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database failed: %w", err)
	}

	db.SetMaxOpenConns(1)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite database failed: %w", err)
	}

	if _, err := db.ExecContext(ctx, `PRAGMA foreign_keys = ON;`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("enable foreign keys failed: %w", err)
	}

	if _, err := db.ExecContext(ctx, schemaSQL); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("initialize sqlite schema failed: %w", err)
	}

	return &SqliteStore{
		db: db,
		q:  generated.New(db),
	}, nil
}

func (s *SqliteStore) Close() error {
	return s.db.Close()
}

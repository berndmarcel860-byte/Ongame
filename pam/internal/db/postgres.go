package db

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Postgres wraps a sql.DB connection pool for PostgreSQL.
type Postgres struct {
	DB *sql.DB
}

func NewPostgres(dsn string) (*Postgres, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	return &Postgres{DB: db}, nil
}

func (p *Postgres) Close() error {
	return p.DB.Close()
}

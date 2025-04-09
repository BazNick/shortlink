package entities

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DB struct {
	Database *sql.DB
}

func NewDB(connection string) *DB {
	db, err := sql.Open("pgx", connection)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		fmt.Println(err)
		return nil
	}

	_, err = db.ExecContext(
		ctx,
		`CREATE TABLE IF NOT EXISTS links (
			short_url varchar(15) NOT NULL,
			original_url text NOT NULL,
			PRIMARY KEY (short_url)
		)`,
	)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return &DB{Database: db}
}

func (db *DB) AddHash(hash, link string) {
	_, err := db.Database.ExecContext(
		context.Background(),
		`INSERT INTO links (short_url, original_url) VALUES ($1, $2)`,
		hash,
		link,
	)
	if err != nil {
		fmt.Println(err)
	}
}

func (db *DB) GetHash(hash string) string {
	row := db.Database.QueryRowContext(
		context.Background(),
		`SELECT original_url FROM links WHERE short_url = $1`,
		hash,
	)

	var link string

	err := row.Scan(&link)
	if err != nil {
		return ""
	}

	return link
}

func (db *DB) CheckValExists(link string) bool {
	row := db.Database.QueryRowContext(
		context.Background(),
		`SELECT original_url FROM links WHERE original_url = $1`,
		link,
	)

	var exists sql.NullString

	err := row.Scan(&exists)

	if err != nil {
		fmt.Println(err)
	}

	if exists.Valid {
		return true
	}

	return false
}

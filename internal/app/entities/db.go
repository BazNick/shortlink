package entities

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
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
			original_url text NOT NULL UNIQUE,
			PRIMARY KEY (short_url)
		)`,
	)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	_, err = db.ExecContext(
		ctx,
		`CREATE INDEX IF NOT EXISTS idx_original_url ON links(original_url);`,
	)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return &DB{Database: db}
}

func (db *DB) AddHash(hash, link string) (string, error) {
	var shortURL string

	err := db.Database.QueryRowContext(
		context.Background(),
		`INSERT INTO links (short_url, original_url) 
		 VALUES ($1, $2) 
		 RETURNING short_url;`,
		hash,
		link,
	).Scan(&shortURL)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			errQueryRow := db.Database.QueryRowContext(
				context.Background(),
				`SELECT short_url FROM links WHERE original_url = $1`,
				link,
			).Scan(&shortURL)

			if errQueryRow != nil {
				return "", fmt.Errorf("conflict, but failed to retrieve short_url: %w", errQueryRow)
			}

			return shortURL, errors.New("conflict")
		}

		return "", err
	}

	return shortURL, nil
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

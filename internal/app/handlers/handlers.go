package handlers

import (
	"database/sql"

	"github.com/BazNick/shortlink/internal/app/entities"
	"github.com/BazNick/shortlink/internal/app/storage"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type (
	JSONLink struct {
		Link string `json:"url"`
	}

	URLHandler struct {
		storage storage.Storage
		path    string
		dbPath  string
		db      *sql.DB
	}

	BatchIn struct {
		CorrelationID string `json:"correlation_id"`
		OriginalURL   string `json:"original_url"`
	}

	BatchOut struct {
		CorrelationID string `json:"correlation_id"`
		ShortURL      string `json:"short_url"`
	}
)

func NewURLHandler(
	storage storage.Storage,
	filePath, dbPath string,
) *URLHandler {
	var db *sql.DB

	if dbStorage, ok := storage.(*entities.DB); ok {
		db = dbStorage.Database
	}

	handler := &URLHandler{
		storage: storage,
		path:    filePath,
		dbPath:  dbPath,
		db:      db,
	}

	return handler
}
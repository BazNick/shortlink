package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/BazNick/shortlink/internal/app/apperr"
	"github.com/BazNick/shortlink/internal/app/entities"
	"github.com/BazNick/shortlink/internal/app/functions"
	"github.com/BazNick/shortlink/internal/app/storage"
	"github.com/gin-gonic/gin"
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

func (handler *URLHandler) AddLink(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		http.Error(c.Writer, apperr.ErrOnlyPOST.Error(), http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		http.Error(c.Writer, apperr.ErrBodyRead.Error(), http.StatusBadRequest)
		return
	}
	c.Request.Body.Close()

	if _, ok := handler.storage.(*entities.DB); !ok {
		alreadyExst := handler.storage.CheckValExists(string(body))
		if alreadyExst {
			http.Error(c.Writer, apperr.ErrLinkExists.Error(), http.StatusBadRequest)
			return
		}
	}

	var (
		randStr  = functions.RandSeq(8)
		hashLink = functions.SchemeAndHost(c.Request) + "/" + randStr
	)

	shortURL, err := handler.storage.AddHash(randStr, string(body), functions.User(c))
	if err != nil {
		if err.Error() == apperr.ErrValAlreadyExists.Error() {
			c.Writer.WriteHeader(http.StatusConflict)
			c.Writer.Write([]byte(functions.SchemeAndHost(c.Request) + "/" + shortURL))
			return
		}
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	c.Writer.Header().Set("content-type", "text/plain")
	c.Writer.WriteHeader(http.StatusCreated)
	c.Writer.Write([]byte(hashLink))
}

func (handler *URLHandler) GetLink(c *gin.Context) {
	if c.Request.Method != http.MethodGet {
		http.Error(c.Writer, apperr.ErrOnlyGET.Error(), http.StatusMethodNotAllowed)
		return
	}

	var (
		id     = c.Param("id")
		pageID = handler.storage.GetHash(id)
	)

	if pageID == "" {
		http.Error(c.Writer, apperr.ErrLinkNotFound.Error(), http.StatusBadRequest)
		return
	}

	c.Writer.Header().Set("Location", pageID)
	c.Writer.Header().Set("Content-Type", "text/html")
	c.Writer.WriteHeader(http.StatusTemporaryRedirect)
}

func (handler *URLHandler) PostJSONLink(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		http.Error(c.Writer, apperr.ErrOnlyPOST.Error(), http.StatusMethodNotAllowed)
		return
	}

	var link JSONLink
	if err := json.NewDecoder(c.Request.Body).Decode(&link); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusBadRequest)
		return
	}

	if _, ok := handler.storage.(*entities.DB); !ok {
		alreadyExst := handler.storage.CheckValExists(link.Link)
		if alreadyExst {
			http.Error(c.Writer, apperr.ErrLinkExists.Error(), http.StatusBadRequest)
			return
		}
	}

	var (
		randStr  = functions.RandSeq(8)
		hashLink = functions.SchemeAndHost(c.Request) + "/" + randStr
	)

	shortURL, err := handler.storage.AddHash(randStr, link.Link, functions.User(c))
	if err != nil {
		if err.Error() == "conflict" {
			resp, err := json.Marshal(map[string]string{
				"result": functions.SchemeAndHost(c.Request) + "/" + shortURL,
			})
			if err != nil {
				http.Error(c.Writer, err.Error(), http.StatusBadRequest)
				return
			}

			c.Writer.Header().Set("content-type", "application/json")
			c.Writer.WriteHeader(http.StatusConflict)
			c.Writer.Write(resp)
			return
		}
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(map[string]string{"result": hashLink})
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusBadRequest)
		return
	}

	c.Writer.Header().Set("content-type", "application/json")
	c.Writer.WriteHeader(http.StatusCreated)
	c.Writer.Write(resp)
}

func (handler *URLHandler) DBPingConn(c *gin.Context) {
	db, err := sql.Open("pgx", handler.dbPath)
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}

	c.Writer.WriteHeader(http.StatusOK)
}

func (handler *URLHandler) BatchLinks(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		http.Error(c.Writer, apperr.ErrOnlyPOST.Error(), http.StatusMethodNotAllowed)
		return
	}

	var links []BatchIn

	if err := json.NewDecoder(c.Request.Body).Decode(&links); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusBadRequest)
		return
	}

	for _, link := range links {
		if handler.storage.CheckValExists(link.OriginalURL) {
			http.Error(c.Writer, apperr.ErrLinkExists.Error(), http.StatusBadRequest)
			return
		}
	}

	out := make([]BatchOut, len(links))

	// если это БД
	if _, ok := handler.storage.(*entities.DB); ok {
		tx, err := handler.db.Begin()
		if err != nil {
			http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		}

		for idx, link := range links {
			shortURL := functions.RandSeq(8)
			_, err := tx.ExecContext(
				context.Background(),
				`INSERT INTO links (short_url, original_url, user_id) VALUES ($1, $2, $3)`,
				shortURL,
				link.OriginalURL,
				functions.User(c),
			)
			if err != nil {
				tx.Rollback()
				http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
			}
			out[idx].CorrelationID = link.CorrelationID
			out[idx].ShortURL = functions.SchemeAndHost(c.Request) + "/" + shortURL

		}

		if err := tx.Commit(); err != nil {
			http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		}
	} else {
		// если это не БД, то сохраняем в файл или в мапу
		for idx, link := range links {
			shortURL := functions.RandSeq(8)
			handler.storage.AddHash(shortURL, link.OriginalURL, functions.User(c))
			out[idx].CorrelationID = link.CorrelationID
			out[idx].ShortURL = functions.SchemeAndHost(c.Request) + "/" + shortURL
		}
	}

	resp, err := json.Marshal(out)

	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusBadRequest)
		return
	}

	c.Writer.Header().Set("content-type", "application/json")
	c.Writer.WriteHeader(http.StatusCreated)
	c.Writer.Write(resp)
}

func (handler *URLHandler) GetUserLinks(c *gin.Context) {
	rows, err := handler.db.QueryContext(
		context.Background(),
		`SELECT short_url, original_url FROM links WHERE user_id = $1`,
		functions.User(c),
	)
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusBadRequest)
	}

	defer rows.Close()

	var result []entities.FileLinks
	for rows.Next() {
		var rec entities.FileLinks
		if err := rows.Scan(&rec.ShortURL, &rec.OriginalURL); err != nil {
			http.Error(c.Writer, err.Error(), http.StatusBadRequest)
		}
		result = append(result, rec)
	}

	if err := rows.Err(); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}

	resp, err := json.Marshal(result)

	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusBadRequest)
		return
	}

	c.Writer.Header().Set("content-type", "application/json")
	c.Writer.Write(resp)
}

package handlers

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
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

	FileLinks struct {
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
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

func NewURLHandler(storage storage.Storage, path, dbPath string) *URLHandler {
	var db *sql.DB

	if dbStorage, ok := storage.(*entities.DB); ok {
		db = dbStorage.Database
	}

	handler := &URLHandler{
		storage: storage,
		path:    path,
		dbPath:  dbPath,
		db:      db,
	}

	// загружаем все ссылки в память из файла
	reader, errOpenFile := os.OpenFile(handler.path, os.O_RDONLY|os.O_CREATE, 0666)
	if errOpenFile != nil {
		log.Fatalf("Ошибка при открытии файла %s: %v", path, errOpenFile)
	}
	defer reader.Close()

	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		var res FileLinks
		err := json.Unmarshal(scanner.Bytes(), &res)
		if err != nil {
			log.Fatalf("Ошибка при открытии файла %s: %v", path, err)
		}
		storage.AddHash(res.ShortURL, res.OriginalURL)
	}
	return handler
}

func (handler *URLHandler) saveToFile(shortURL, originalURL string) error {
	file, err := os.OpenFile(handler.path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	bufWriter := bufio.NewWriter(file)

	data, err := json.Marshal(FileLinks{
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	})
	if err != nil {
		file.Close()
		return err
	}

	data = append(data, '\n')

	if _, err = bufWriter.Write(data); err != nil {
		file.Close()
		return err
	}

	if err := bufWriter.Flush(); err != nil {
		file.Close()
		return err
	}

	if err := file.Sync(); err != nil {
		file.Close()
		return err
	}

	return file.Close()
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

	shortURL, err := handler.storage.AddHash(randStr, string(body))
	if err != nil {
		if err.Error() == "conflict" {
			c.Writer.WriteHeader(http.StatusConflict)
			c.Writer.Write([]byte(functions.SchemeAndHost(c.Request) + "/" + shortURL))
			return
		}
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := handler.saveToFile(randStr, string(body)); err != nil {
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

	shortURL, err := handler.storage.AddHash(randStr, link.Link)
	if err != nil {
		if err.Error() == "conflict" {
			c.Writer.Header().Set("content-type", "application/json")
			c.Writer.WriteHeader(http.StatusConflict)
			c.Writer.Write([]byte(functions.SchemeAndHost(c.Request) + "/" + shortURL))
			return
		}
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := handler.saveToFile(randStr, link.Link); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusBadRequest)
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
				`INSERT INTO links (short_url, original_url) VALUES ($1, $2)`,
				shortURL,
				link.OriginalURL,
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
	}

	// если это не БД, то сохраняем в файл и в мапу
	if _, ok := handler.storage.(*entities.HashDict); ok {
		for idx, link := range links {
			shortURL := functions.RandSeq(8)
			handler.storage.AddHash(functions.RandSeq(8), link.OriginalURL)
			if err := handler.saveToFile(functions.RandSeq(8), link.OriginalURL); err != nil {
				http.Error(c.Writer, err.Error(), http.StatusBadRequest)
				return
			}
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

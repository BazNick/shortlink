package handlers

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/BazNick/shortlink/internal/app/apperr"
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
		db      string
	}
)

func NewURLHandler(storage storage.Storage, path, db string) *URLHandler {
	handler := &URLHandler{
		storage: storage,
		path:    path,
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
		http.Error(c.Writer, apperr.ErrOnlyPOST, http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		http.Error(c.Writer, apperr.ErrBodyRead, http.StatusBadRequest)
		return
	}
	c.Request.Body.Close()

	alreadyExst := handler.storage.CheckValExists(string(body))
	if alreadyExst {
		http.Error(c.Writer, apperr.ErrLinkExists, http.StatusBadRequest)
		return
	}

	var scheme string
	if c.Request.TLS != nil {
		scheme = "https://"
	} else {
		scheme = "http://"
	}

	var (
		randStr  = functions.RandSeq(8)
		hashLink = scheme + c.Request.Host + "/" + randStr
	)

	handler.storage.AddHash(randStr, string(body))
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
		http.Error(c.Writer, apperr.ErrOnlyGET, http.StatusMethodNotAllowed)
		return
	}

	var (
		id     = c.Param("id")
		pageID = handler.storage.GetHash(id)
	)

	if pageID == "" {
		http.Error(c.Writer, apperr.ErrLinkNotFound, http.StatusBadRequest)
		return
	}

	c.Writer.Header().Set("Location", pageID)
	c.Writer.Header().Set("Content-Type", "text/html")
	c.Writer.WriteHeader(http.StatusTemporaryRedirect)
}

func (handler *URLHandler) PostJSONLink(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		http.Error(c.Writer, apperr.ErrOnlyPOST, http.StatusMethodNotAllowed)
		return
	}

	var link JSONLink
	if err := json.NewDecoder(c.Request.Body).Decode(&link); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusBadRequest)
		return
	}

	if handler.storage.CheckValExists(link.Link) {
		http.Error(c.Writer, apperr.ErrLinkExists, http.StatusBadRequest)
		return
	}

	var scheme string
	if c.Request.TLS != nil {
		scheme = "https://"
	} else {
		scheme = "http://"
	}

	var (
		randStr  = functions.RandSeq(8)
		hashLink = scheme + c.Request.Host + "/" + randStr
	)

	handler.storage.AddHash(randStr, link.Link)
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
	ps := fmt.Sprintf(handler.db)
	
	db, err := sql.Open("pgx", ps)
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

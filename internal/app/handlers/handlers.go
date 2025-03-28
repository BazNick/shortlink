package handlers

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/BazNick/shortlink/internal/app/apperr"
	"github.com/BazNick/shortlink/internal/app/functions"
	"github.com/BazNick/shortlink/internal/app/storage"
	"github.com/gin-gonic/gin"
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
	}
)

func NewURLHandler(storage storage.Storage, path string) *URLHandler {
	handler := &URLHandler{storage: storage, path: path}

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

	alreadyExst := handler.storage.CheckValExists(link.Link)
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

	handler.storage.AddHash(randStr, link.Link)

	// пишем ссылку в файл
	writer, errFile := os.OpenFile(handler.path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if errFile != nil {
		http.Error(c.Writer, errFile.Error(), http.StatusBadRequest)
		return
	}
	defer writer.Close()

	data, errMarshal := json.Marshal(FileLinks{ShortURL: randStr, OriginalURL: link.Link})
	if errMarshal != nil {
		http.Error(c.Writer, errMarshal.Error(), http.StatusBadRequest)
		return
	}

	data = append(data, '\n')

	_, errWriteFile := writer.Write(data)
	if errWriteFile != nil {
		http.Error(c.Writer, errWriteFile.Error(), http.StatusBadRequest)
		return
	}

	resp, err := json.Marshal(map[string]string{"result": hashLink})
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusBadRequest)
		return
	}

	c.Writer.Header().Set("content-type", "application/json")
	c.Writer.WriteHeader(http.StatusCreated)
	c.Writer.Write([]byte(resp))
}

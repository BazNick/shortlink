package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/BazNick/shortlink/internal/app/apperr"
	"github.com/BazNick/shortlink/internal/app/functions"
	"github.com/BazNick/shortlink/internal/app/storage"
	"github.com/gin-gonic/gin"
)

type JSONLink struct {
	Link string `json:"url"`
}

type URLHandler struct {
	storage storage.Storage
}

func NewURLHandler(storage storage.Storage) *URLHandler {
	return &URLHandler{storage: storage}
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

	resp, err := json.Marshal(map[string]string{"result": hashLink})
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusBadRequest)
		return
	}

	c.Writer.Header().Set("content-type", "application/json")
	c.Writer.WriteHeader(http.StatusCreated)
	c.Writer.Write([]byte(resp))
}

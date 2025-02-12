package handlers

import (
	"io"
	"net/http"

	"github.com/BazNick/shortlink/internal/app/apperr"
	"github.com/BazNick/shortlink/internal/app/entities"
	"github.com/BazNick/shortlink/internal/app/functions"
	"github.com/gin-gonic/gin"
)

func AddLink(c *gin.Context) {
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

	alreadyExst := entities.CheckValExists(entities.Hash, string(body))
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

	entities.Hash.AddHash(randStr, string(body))

	c.Writer.Header().Set("content-type", "text/plain")
	c.Writer.WriteHeader(http.StatusCreated)
	c.Writer.Write([]byte(hashLink))
}

func GetLink(c *gin.Context) {
	if c.Request.Method != http.MethodGet {
		http.Error(c.Writer, apperr.ErrOnlyGET, http.StatusMethodNotAllowed)
		return
	}

	var (
		id     = c.Param("id")
		exists = entities.Hash.GetHash(id)
	)

	if exists == "" {
		http.Error(c.Writer, apperr.ErrLinkNotFound, http.StatusBadRequest)
		return
	}

	c.Writer.Header().Set("Location", exists)
	c.Writer.WriteHeader(http.StatusTemporaryRedirect)
}

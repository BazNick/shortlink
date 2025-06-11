package handlers

import (
	"io"
	"net/http"

	"github.com/BazNick/shortlink/internal/app/apperr"
	"github.com/BazNick/shortlink/internal/app/entities"
	"github.com/BazNick/shortlink/internal/app/functions"
	"github.com/gin-gonic/gin"
)

func (handler *URLHandler) AddLink(c *gin.Context) {
	user, err := functions.GetUser(c)
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusBadRequest)
		return
	}

	if c.Request.Method != http.MethodPost {
		http.Error(c.Writer, apperr.ErrOnlyPOST.Error(), http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		http.Error(c.Writer, apperr.ErrBodyRead.Error(), http.StatusBadRequest)
		return
	}
	defer c.Request.Body.Close()

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

	shortURL, err := handler.storage.AddHash(randStr, string(body), user)
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

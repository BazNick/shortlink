package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/BazNick/shortlink/internal/app/apperr"
	"github.com/BazNick/shortlink/internal/app/entities"
	"github.com/BazNick/shortlink/internal/app/functions"
	"github.com/gin-gonic/gin"
)

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
	userID, err := functions.User(c, false)
	if err != nil {
		http.Error(c.Writer, "unauthorized", http.StatusUnauthorized)
		return
	}
	shortURL, err := handler.storage.AddHash(randStr, link.Link, userID)
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
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/BazNick/shortlink/internal/app/entities"
	"github.com/BazNick/shortlink/internal/app/functions"
	"github.com/gin-gonic/gin"
)

func (handler *URLHandler) DeleteUserLinks(c *gin.Context) {
	user, err := functions.GetUser(c)
	if err != nil {
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	var links []string
	if err := json.NewDecoder(c.Request.Body).Decode(&links); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusBadRequest)
		return
	}

	entities.DeleteChan <- entities.DeleteRequest{
		UserID:    user,
		ShortURLs: links,
	}

	c.Writer.WriteHeader(http.StatusAccepted)
}

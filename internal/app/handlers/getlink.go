package handlers

import (
	"net/http"

	"github.com/BazNick/shortlink/internal/app/apperr"
	"github.com/gin-gonic/gin"
)

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
		http.Error(c.Writer, apperr.ErrLinkNotFound.Error(), http.StatusGone)
		return
	}

	c.Writer.Header().Set("Location", pageID)
	c.Writer.Header().Set("Content-Type", "text/html")
	c.Writer.WriteHeader(http.StatusTemporaryRedirect)
}
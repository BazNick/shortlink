package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/BazNick/shortlink/internal/app/apperr"
	"github.com/BazNick/shortlink/internal/app/entities"
	"github.com/BazNick/shortlink/internal/app/functions"
	"github.com/gin-gonic/gin"
)

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
	userID, err := functions.User(c, false)
	if err != nil {
		http.Error(c.Writer, "unauthorized", http.StatusUnauthorized)
		return
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
				userID,
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
			handler.storage.AddHash(shortURL, link.OriginalURL, userID)
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
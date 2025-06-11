package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/BazNick/shortlink/internal/app/entities"
	"github.com/BazNick/shortlink/internal/app/functions"
	"github.com/gin-gonic/gin"
)

func (handler *URLHandler) GetUserLinks(c *gin.Context) {
	userID, err := functions.User(c, true)
	if err != nil {
		http.Error(c.Writer, "unauthorized", http.StatusBadRequest)
		return
	}
	if userID == "" {
		http.Error(c.Writer, "unauthorized", http.StatusUnauthorized)
		return
	}
	rows, err := handler.db.QueryContext(
		context.Background(),
		`SELECT short_url, original_url FROM links WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusBadRequest)
	}

	defer rows.Close()

	var result []entities.FileLinks
	for rows.Next() {
		var rec entities.FileLinks
		if err := rows.Scan(&rec.ShortURL, &rec.OriginalURL); err != nil {
			http.Error(c.Writer, err.Error(), http.StatusBadRequest)
		}
		result = append(result, rec)
	}

	if len(result) == 0 {
		c.Writer.WriteHeader(http.StatusNoContent)
		return
	}

	if err := rows.Err(); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}

	resp, err := json.Marshal(result)

	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusBadRequest)
		return
	}

	c.Writer.Header().Set("content-type", "application/json")
	c.Writer.Write(resp)
}
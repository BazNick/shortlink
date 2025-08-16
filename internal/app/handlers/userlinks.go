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
	user, err := functions.GetUser(c)
	if err != nil {
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	rows, err := handler.db.QueryContext(
		context.Background(),
		`SELECT short_url, original_url FROM links WHERE user_id = $1`,
		user,
	)
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	baseURL := functions.SchemeAndHost(c.Request)
	result := make([]entities.FileLinks, 0, 10)

	for rows.Next() {
		var rec entities.FileLinks
		if err := rows.Scan(&rec.ShortURL, &rec.OriginalURL); err != nil {
			http.Error(c.Writer, err.Error(), http.StatusBadRequest)
			return
		}
		rec.ShortURL = baseURL + "/" + rec.ShortURL
		result = append(result, rec)
	}

	if len(result) == 0 {
		c.Writer.WriteHeader(http.StatusNoContent)
		return
	}

	if err := rows.Err(); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(result)

	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusBadRequest)
		return
	}

	c.Writer.Header().Set("content-type", "application/json")
	c.Writer.Write(resp)
}

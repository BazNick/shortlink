package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// DBPingConn обрабатывает HTTP-запросы GET для тестирования доступности БД.
// Пример:
//
//	GET /ping
//
//	Ответ: 200 OK (если доступ к базе данных возможен)
//	Ответ: 500 Internal Server Error (если доступ к базе данных невозможен)
func (handler *URLHandler) DBPingConn(c *gin.Context) {
	db, err := sql.Open("pgx", handler.dbPath)
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}

	c.Writer.WriteHeader(http.StatusOK)
}

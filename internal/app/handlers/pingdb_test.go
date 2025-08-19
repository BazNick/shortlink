package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/BazNick/shortlink/internal/app/entities"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDBPingConn(t *testing.T) {
	storage := entities.NewHashDict()
	handler := NewURLHandler(
		storage,
		"test.json",
		"postgres://user:password@localhost:5432/dbname",
	)

	router := gin.Default()
	router.GET("/ping", handler.DBPingConn)

	request := httptest.NewRequest(http.MethodGet, "http://localhost:8080/ping", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, request)

	assert.NotPanics(t, func() {
		router.ServeHTTP(w, request)
	})
}

func TestDBPingConn_InvalidDBPath(t *testing.T) {
	storage := entities.NewHashDict()
	handler := NewURLHandler(
		storage,
		"test.json",
		"invalid://connection/string",
	)

	router := gin.Default()
	router.GET("/ping", handler.DBPingConn)

	request := httptest.NewRequest(http.MethodGet, "http://localhost:8080/ping", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, request)

	assert.NotEqual(t, http.StatusOK, w.Code)
}

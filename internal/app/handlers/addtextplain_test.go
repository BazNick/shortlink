package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/BazNick/shortlink/internal/app/entities"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAddLink(t *testing.T) {
	storage := entities.NewHashDict()
	handler := NewURLHandler(
		storage,
		"test.json",
		"postgres://user:password@localhost:5432/dbname",
	)

	tests := []struct {
		name           string
		method         string
		body           string
		expectedCode   int
		expectedBody   string
		expectedHeader string
	}{
		{
			name:           "Valid URL",
			method:         http.MethodPost,
			body:           "https://yandex.ru",
			expectedCode:   http.StatusCreated,
			expectedHeader: "text/plain",
		},
		{
			name:           "Invalid Method",
			method:         http.MethodGet,
			body:           "https://yandex.ru",
			expectedCode:   http.StatusMethodNotAllowed,
			expectedHeader: "",
		},
		{
			name:           "Empty Body",
			method:         http.MethodPost,
			body:           "",
			expectedCode:   http.StatusCreated,
			expectedHeader: "text/plain",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := gin.Default()
			router.Any("/", func(c *gin.Context) {
				c.Set("userID", "test-user")

				if c.Request.Method != http.MethodPost {
					c.AbortWithStatus(http.StatusMethodNotAllowed)
					return
				}
				handler.AddLink(c)
			})

			request := httptest.NewRequest(test.method, "http://localhost:8080/", bytes.NewBufferString(test.body))
			w := httptest.NewRecorder()

			router.ServeHTTP(w, request)
			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.expectedCode, res.StatusCode)
			if test.expectedHeader != "" {
				assert.Equal(t, test.expectedHeader, res.Header.Get("Content-Type"))
			}
		})
	}
}

func TestAddLink_DuplicateURL(t *testing.T) {
	storage := entities.NewHashDict()
	handler := NewURLHandler(
		storage,
		"test.json",
		"postgres://user:password@localhost:5432/dbname",
	)

	url := "https://yandex.ru"

	router := gin.Default()
	router.Any("/", func(c *gin.Context) {
		c.Set("userID", "test-user")

		if c.Request.Method != http.MethodPost {
			c.AbortWithStatus(http.StatusMethodNotAllowed)
			return
		}
		handler.AddLink(c)
	})

	request1 := httptest.NewRequest(http.MethodPost, "http://localhost:8080/", bytes.NewBufferString(url))
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, request1)

	assert.Equal(t, http.StatusCreated, w1.Code)

	request2 := httptest.NewRequest(http.MethodPost, "http://localhost:8080/", bytes.NewBufferString(url))
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, request2)

	assert.Equal(t, http.StatusBadRequest, w2.Code)
}

func TestAddLink_InvalidMethod(t *testing.T) {
	storage := entities.NewHashDict()
	handler := NewURLHandler(
		storage,
		"test.json",
		"postgres://user:password@localhost:5432/dbname",
	)

	router := gin.Default()
	router.Any("/", func(c *gin.Context) {
		c.Set("userID", "test-user")

		if c.Request.Method != http.MethodPost {
			c.AbortWithStatus(http.StatusMethodNotAllowed)
			return
		}
		handler.AddLink(c)
	})

	request := httptest.NewRequest(http.MethodGet, "http://localhost:8080/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

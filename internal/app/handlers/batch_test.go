package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/BazNick/shortlink/internal/app/entities"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestBatchLinks(t *testing.T) {
	storage := entities.NewHashDict()
	handler := NewURLHandler(
		storage,
		"test.json",
		"postgres://user:password@localhost:5432/dbname",
	)

	tests := []struct {
		name         string
		method       string
		links        []BatchIn
		expectedCode int
	}{
		{
			name:   "Valid batch request",
			method: http.MethodPost,
			links: []BatchIn{
				{CorrelationID: "1", OriginalURL: "https://example1.com"},
				{CorrelationID: "2", OriginalURL: "https://example2.com"},
			},
			expectedCode: http.StatusCreated,
		},
		{
			name:   "Invalid method",
			method: http.MethodGet,
			links: []BatchIn{
				{CorrelationID: "1", OriginalURL: "https://example1.com"},
			},
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "Empty batch",
			method:       http.MethodPost,
			links:        []BatchIn{},
			expectedCode: http.StatusCreated,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := gin.Default()
			router.Any("/batch", func(c *gin.Context) {
				c.Set("userID", "test-user")

				if c.Request.Method != http.MethodPost {
					c.AbortWithStatus(http.StatusMethodNotAllowed)
					return
				}
				handler.BatchLinks(c)
			})

			jsonData, _ := json.Marshal(test.links)
			request := httptest.NewRequest(test.method, "http://localhost:8080/batch", bytes.NewBuffer(jsonData))
			request.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, request)
			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.expectedCode, res.StatusCode)
		})
	}
}

func TestBatchLinks_DuplicateURL(t *testing.T) {
	storage := entities.NewHashDict()
	handler := NewURLHandler(
		storage,
		"test.json",
		"postgres://user:password@localhost:5432/dbname",
	)

	storage.AddHash("abc123", "https://example.com", "user1")

	links := []BatchIn{
		{CorrelationID: "1", OriginalURL: "https://example.com"},
		{CorrelationID: "2", OriginalURL: "https://example2.com"},
	}

	router := gin.Default()
	router.Any("/batch", func(c *gin.Context) {
		c.Set("userID", "test-user")

		if c.Request.Method != http.MethodPost {
			c.AbortWithStatus(http.StatusMethodNotAllowed)
			return
		}
		handler.BatchLinks(c)
	})

	jsonData, _ := json.Marshal(links)
	request := httptest.NewRequest(http.MethodPost, "http://localhost:8080/batch", bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBatchLinks_InvalidJSON(t *testing.T) {
	storage := entities.NewHashDict()
	handler := NewURLHandler(
		storage,
		"test.json",
		"postgres://user:password@localhost:5432/dbname",
	)

	router := gin.Default()
	router.Any("/batch", func(c *gin.Context) {
		c.Set("userID", "test-user")

		if c.Request.Method != http.MethodPost {
			c.AbortWithStatus(http.StatusMethodNotAllowed)
			return
		}
		handler.BatchLinks(c)
	})

	request := httptest.NewRequest(http.MethodPost, "http://localhost:8080/batch", bytes.NewBufferString("invalid json"))
	request.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBatchLinks_ResponseFormat(t *testing.T) {
	storage := entities.NewHashDict()
	handler := NewURLHandler(
		storage,
		"test.json",
		"postgres://user:password@localhost:5432/dbname",
	)

	links := []BatchIn{
		{CorrelationID: "1", OriginalURL: "https://example1.com"},
		{CorrelationID: "2", OriginalURL: "https://example2.com"},
	}

	router := gin.Default()
	router.Any("/batch", func(c *gin.Context) {
		c.Set("userID", "test-user")

		if c.Request.Method != http.MethodPost {
			c.AbortWithStatus(http.StatusMethodNotAllowed)
			return
		}
		handler.BatchLinks(c)
	})

	jsonData, _ := json.Marshal(links)
	request := httptest.NewRequest(http.MethodPost, "http://localhost:8080/batch", bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response []BatchOut
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)
	assert.Equal(t, "1", response[0].CorrelationID)
	assert.Equal(t, "2", response[1].CorrelationID)
}

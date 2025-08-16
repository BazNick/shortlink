package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/BazNick/shortlink/internal/app/entities"
	"github.com/BazNick/shortlink/internal/app/functions"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetLink(t *testing.T) {
	storage := entities.NewHashDict()
	handler := NewURLHandler(
		storage,
		"test.json",
		"postgres://user:password@localhost:5432/dbname",
	)

	var (
		randomStr   = functions.RandSeq(8)
		originalURL = "https://yandex.ru"
		userID      = "test"
	)
	storage.AddHash(randomStr, originalURL, userID)

	type want struct {
		method         string
		shortURL       string
		expectedCode   int
		expectedHeader string
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "Valid URL",
			want: want{
				method:         http.MethodGet,
				shortURL:       randomStr,
				expectedCode:   http.StatusTemporaryRedirect,
				expectedHeader: originalURL,
			},
		},
		{
			name: "Invalid URL",
			want: want{
				method:         http.MethodGet,
				shortURL:       "nonexistent",
				expectedCode:   http.StatusGone,
				expectedHeader: "",
			},
		},
		{
			name: "Invalid Method",
			want: want{
				method:         http.MethodPost,
				shortURL:       randomStr,
				expectedCode:   http.StatusMethodNotAllowed,
				expectedHeader: "",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := gin.Default()
			router.Any("/:id", func(c *gin.Context) {
				if c.Request.Method != http.MethodGet {
					c.AbortWithStatus(http.StatusMethodNotAllowed)
					return
				}
				handler.GetLink(c)
			})

			request := httptest.NewRequest(test.want.method, "http://localhost:8080/"+test.want.shortURL, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, request)
			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.expectedCode, res.StatusCode)
			if test.want.expectedHeader != "" {
				assert.Equal(t, test.want.expectedHeader, res.Header.Get("Location"))
			}
		})
	}
}

func BenchmarkGetLink_Valid(b *testing.B) {
	storage := entities.NewHashDict()
	handler := NewURLHandler(
		storage,
		"test.json",
		"postgres://user:password@localhost:5432/dbname",
	)

	randomStr := functions.RandSeq(8)
	originalURL := "https://yandex.ru"
	userID := "test"
	storage.AddHash(randomStr, originalURL, userID)

	router := gin.Default()
	router.Any("/:id", func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			c.AbortWithStatus(http.StatusMethodNotAllowed)
			return
		}
		handler.GetLink(c)
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request := httptest.NewRequest(http.MethodGet, "http://localhost:8080/"+randomStr, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, request)
	}
}

func BenchmarkGetLink_Invalid(b *testing.B) {
	storage := entities.NewHashDict()
	handler := NewURLHandler(
		storage,
		"test.json",
		"postgres://user:password@localhost:5432/dbname",
	)

	router := gin.Default()
	router.Any("/:id", func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			c.AbortWithStatus(http.StatusMethodNotAllowed)
			return
		}
		handler.GetLink(c)
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request := httptest.NewRequest(http.MethodGet, "http://localhost:8080/nonexistent", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, request)
	}
}

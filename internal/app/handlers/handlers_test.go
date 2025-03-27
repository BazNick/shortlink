package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/BazNick/shortlink/internal/app/entities"
	"github.com/BazNick/shortlink/internal/app/functions"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddLink(t *testing.T) {
	storage := entities.NewHashDict()
	handler := NewURLHandler(storage, "test.json")

	type want struct {
		method       string
		url          string
		expectedCode int
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "Test POST success",
			want: want{
				method:       http.MethodPost,
				url:          "https://yandex.ru",
				expectedCode: http.StatusCreated,
			},
		},
		{
			name: "Test POST fail (wrong method)",
			want: want{
				method:       http.MethodGet,
				url:          "https://yandex.ru",
				expectedCode: http.StatusNotFound,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := gin.Default()
			router.POST("/", handler.AddLink)

			data := strings.NewReader(test.want.url)
			request := httptest.NewRequest(test.want.method, "http://localhost:8080/", data)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, request)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.expectedCode, res.StatusCode)

			if res.StatusCode == http.StatusCreated {
				body, err := io.ReadAll(res.Body)
				require.NoError(t, err)

				resBody := string(body)
				assert.Contains(t, resBody, "http://localhost:8080/")
			}
		})
	}
}

func TestGetLink(t *testing.T) {
	storage := entities.NewHashDict()
	handler := NewURLHandler(storage, "test.json")

	randomStr := functions.RandSeq(8)
	originalURL := "https://yandex.ru"
	storage.AddHash(randomStr, originalURL)

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
				expectedCode:   http.StatusBadRequest,
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

func TestPostJSONLink(t *testing.T) {
	storage := entities.NewHashDict()
	handler := NewURLHandler(storage, "test.json")

	type want struct {
		method       string
		body         string
		expectedCode int
		expectResult bool
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "Test POST JSON success",
			want: want{
				method:       http.MethodPost,
				body:         `{"url": "https://yandex.ru"}`,
				expectedCode: http.StatusCreated,
				expectResult: true,
			},
		},
		{
			name: "Test POST fail (invalid JSON)",
			want: want{
				method:       http.MethodPost,
				body:         `{"url":}`,
				expectedCode: http.StatusBadRequest,
				expectResult: false,
			},
		},
		{
			name: "Test POST fail (duplicate URL)",
			want: want{
				method:       http.MethodPost,
				body:         `{"url": "https://duplicate.ru"}`,
				expectedCode: http.StatusCreated,
				expectResult: true,
			},
		},
		{
			name: "Test POST fail (URL already exists)",
			want: want{
				method:       http.MethodPost,
				body:         `{"url": "https://duplicate.ru"}`,
				expectedCode: http.StatusBadRequest,
				expectResult: false,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := gin.Default()
			router.POST("/", handler.PostJSONLink)

			data := strings.NewReader(test.want.body)
			request := httptest.NewRequest(test.want.method, "http://localhost:8080/", data)
			request.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, request)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.expectedCode, res.StatusCode)

			if test.want.expectResult {
				body, err := io.ReadAll(res.Body)
				require.NoError(t, err)

				var resBody map[string]string
				err = json.Unmarshal(body, &resBody)
				require.NoError(t, err)

				assert.Contains(t, resBody["result"], "http://localhost:8080/")
			}
		})
	}
}
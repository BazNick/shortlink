package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/BazNick/shortlink/internal/app/entities"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddLink(t *testing.T) {
	storage := entities.NewHashDict()
	handler := NewURLHandler(
		storage,
		"test.json",
		"postgres://user:password@localhost:5432/dbname",
	)

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

			if res.StatusCode != 401 {
				assert.Equal(t, test.want.expectedCode, res.StatusCode)

				if res.StatusCode == http.StatusCreated {
					body, err := io.ReadAll(res.Body)
					require.NoError(t, err)

					resBody := string(body)
					assert.Contains(t, resBody, "http://localhost:8080/")
				}
			}
		})
	}
}
package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/BazNick/shortlink/cmd/middleware/auth"
	"github.com/BazNick/shortlink/internal/app/entities"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostJSONLink(t *testing.T) {
	storage := entities.NewHashDict()
	handler := NewURLHandler(
		storage,
		"test.json",
		"postgres://user:password@localhost:5432/dbname",
	)

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
				body:         `{"url": "https://vk.ru"}`,
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
			name: "Test POST fail (duplicate1 URL)",
			want: want{
				method:       http.MethodPost,
				body:         `{"url": "https://duplicate1.ru"}`,
				expectedCode: http.StatusCreated,
				expectResult: true,
			},
		},
		{
			name: "Test POST fail (URL already exists)",
			want: want{
				method:       http.MethodPost,
				body:         `{"url": "https://duplicate1.ru"}`,
				expectedCode: http.StatusBadRequest,
				expectResult: false,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := gin.Default()
			router.Use(auth.Auth())
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

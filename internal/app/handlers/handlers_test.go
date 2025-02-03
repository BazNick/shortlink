package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/BazNick/shortlink/internal/app/apperr"
	"github.com/BazNick/shortlink/internal/app/entities"
	"github.com/BazNick/shortlink/internal/app/functions"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAddLink(t *testing.T) {
	type want struct {
		method       string
		url          string
		expectedCode int
		expectedBody string
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
				expectedBody: "http://localhost:8080/" + functions.RandSeq(8),
			},
		},
		{
			name: "Test POST fail",
			want: want{
				method:       http.MethodGet,
				url:          "https://yandex.ru",
				expectedCode: http.StatusBadRequest,
				expectedBody: "",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := gin.Default()
			router.POST("/", AddLink)

			data := strings.NewReader(test.want.url)
			request := httptest.NewRequest(http.MethodPost, "http://localhost:8080/", data)

			w := httptest.NewRecorder()

			router.ServeHTTP(w, request)

			res := w.Result()

			assert.Equal(t, test.want.expectedCode, res.StatusCode)

			defer res.Body.Close()
		})
	}
}

func TestGetLink(t *testing.T) {
	hashDict := make(entities.HashDict, 1)
	randomStr := functions.RandSeq(8)
	hashDict.AddHash(randomStr, "https://yandex.ru")

	type want struct {
		method         string
		shortURL       string
		expectedCode   int
		expectedHeader string
		expectedError  string
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "Invalid URL",
			want: want{
				method:         http.MethodGet,
				shortURL:       `"12345test"`,
				expectedCode:   400,
				expectedHeader: "",
				expectedError:  apperr.ErrLinkNotFound,
			},
		},
		{
			name: "Invalid Method",
			want: want{
				method:         http.MethodPost,
				shortURL:       `"12345test"`,
				expectedCode:   400,
				expectedHeader: "",
				expectedError:  apperr.ErrOnlyGET,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := gin.Default()
			router.GET("/:id", GetLink)

			request := httptest.NewRequest(http.MethodGet, "http://localhost:8080/"+test.want.shortURL, nil)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, request)
			res := w.Result()

			assert.Equal(t, test.want.expectedCode, res.StatusCode)
			defer res.Body.Close()

			assert.Equal(t, test.want.expectedHeader, res.Header.Get("Location"))
		})
	}
}

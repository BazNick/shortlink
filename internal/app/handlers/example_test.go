package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/BazNick/shortlink/internal/app/entities"
	"github.com/BazNick/shortlink/internal/app/handlers"
	"github.com/gin-gonic/gin"
)

func ExampleNewURLHandler() {
	storage := entities.NewHashDict()

	handler := handlers.NewURLHandler(storage, "urls.json", "postgres://user:pass@localhost/db")

	fmt.Printf("Handler created: %v\n", handler != nil)

	// Output:
	// Handler created: true
}

func ExampleURLHandler_PostJSONLink() {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	storage := entities.NewHashDict()
	handler := handlers.NewURLHandler(storage, "urls.json", "")

	router.POST("/api/shorten", func(c *gin.Context) {
		c.Set("userID", "test-user")
		handler.PostJSONLink(c)
	})

	requestBody := handlers.JSONLink{
		Link: "https://example.com/very/long/url",
	}
	jsonData, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", "/api/shorten", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	fmt.Printf("Status: %d\n", w.Code)
	response := w.Body.String()
	if strings.Contains(response, `"result":"http:///`) {
		fmt.Printf("Response: {\"result\":\"http:///HASH\"}\n")
	} else {
		fmt.Printf("Response: %s\n", response)
	}

	// Output:
	// Status: 201
	// Response: {"result":"http:///HASH"}
}

func ExampleURLHandler_GetLink() {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	storage := entities.NewHashDict()
	handler := handlers.NewURLHandler(storage, "urls.json", "")

	storage.AddHash("abc123", "https://example.com/very/long/url", "user123")

	router.GET("/:id", handler.GetLink)

	req, _ := http.NewRequest("GET", "/abc123", nil)

	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Location: %s\n", w.Header().Get("Location"))

	// Output:
	// Status: 307
	// Location: https://example.com/very/long/url
}

func ExampleJSONLink() {
	link := handlers.JSONLink{
		Link: "https://example.com/very/long/url",
	}

	jsonData, _ := json.Marshal(link)
	fmt.Printf("JSON: %s\n", string(jsonData))

	// Output:
	// JSON: {"url":"https://example.com/very/long/url"}
}

func ExampleBatchIn() {
	batchItem := handlers.BatchIn{
		CorrelationID: "req-123",
		OriginalURL:   "https://example.com/very/long/url",
	}

	jsonData, _ := json.Marshal(batchItem)
	fmt.Printf("Batch input: %s\n", string(jsonData))

	// Output:
	// Batch input: {"correlation_id":"req-123","original_url":"https://example.com/very/long/url"}
}

func ExampleBatchOut() {
	batchResponse := handlers.BatchOut{
		CorrelationID: "req-123",
		ShortURL:      "https://shortener.com/abc123",
	}

	jsonData, _ := json.Marshal(batchResponse)
	fmt.Printf("Batch output: %s\n", string(jsonData))

	// Output:
	// Batch output: {"correlation_id":"req-123","short_url":"https://shortener.com/abc123"}
}

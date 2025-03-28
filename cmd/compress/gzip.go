package compress

import (
	"compress/gzip"
	"io"
	"strings"
	"net/http"

	"github.com/gin-gonic/gin"
)

type gzipWriter struct {
	gin.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func GzipHandle() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Content-Encoding") == "gzip" {
			gz, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				c.String(http.StatusBadRequest, "Ошибка декомпрессии запроса")
				c.Abort()
				return
			}
			defer gz.Close()
			c.Request.Body = gz
		}

		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		c.Next()
		contentType := c.Writer.Header().Get("Content-Type")
		if !(strings.HasPrefix(contentType, "application/json") || strings.HasPrefix(contentType, "text/html")) {
			return
		}

		gz := gzip.NewWriter(c.Writer)
		defer gz.Close()

		c.Writer.Header().Set("Content-Encoding", "gzip")
		c.Writer = gzipWriter{ResponseWriter: c.Writer, Writer: gz}
	}
}

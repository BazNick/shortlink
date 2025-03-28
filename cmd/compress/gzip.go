package compress

import (
	"compress/gzip"
	"io"
	"strings"

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
	return gin.HandlerFunc(func(c *gin.Context) {
		if c.GetHeader("Accept-Encoding") == "gzip" {

			contentType := c.GetHeader("Content-Type")
			if !(strings.HasPrefix(contentType, "application/json") || strings.HasPrefix(contentType, "text/html")) {
				return
			}

			gz, err := gzip.NewWriterLevel(c.Writer, gzip.BestSpeed)
			if err != nil {
				io.WriteString(c.Writer, err.Error())
				return
			}
			defer gz.Close()

			c.Writer.Header().Set("Content-Encoding", "gzip")

			c.Writer = gzipWriter{ResponseWriter: c.Writer, Writer: gz}

			c.Next()
		}

		if c.Writer.Status() >= 300 && c.Writer.Status() < 400 {
			return
		}

		c.Next()
	})
}

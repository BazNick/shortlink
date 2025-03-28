package compress

import (
	"compress/gzip"
	"strings"
	"net/http"

	"github.com/gin-gonic/gin"
)

type gzipWriter struct {
    gin.ResponseWriter
    writer *gzip.Writer
}

func (g *gzipWriter) Write(data []byte) (int, error) {
    return g.writer.Write(data)
}

func GzipHandle() gin.HandlerFunc {
    return func(c *gin.Context) {
        if strings.Contains(c.GetHeader("Content-Encoding"), "gzip") {
            gz, err := gzip.NewReader(c.Request.Body)
            if err != nil {
                c.AbortWithError(http.StatusBadRequest, err)
                return
            }
            defer gz.Close()
            c.Request.Body = gz
        }

        if strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
            gz := gzip.NewWriter(c.Writer)
            defer gz.Close()

            originalWriter := c.Writer
            c.Writer = &gzipWriter{
                ResponseWriter: originalWriter,
                writer:        gz,
            }

            c.Header("Content-Encoding", "gzip")
            c.Header("Vary", "Accept-Encoding")
            c.Header("Content-Type", c.Writer.Header().Get("Content-Type"))
            c.Next()

            return
        }
        c.Next()
    }
}
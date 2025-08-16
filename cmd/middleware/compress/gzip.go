package compress

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type gzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

// Write сжимает и записывает данные с использованием gzip-компрессии.
//
// Параметры:
//   - data: срез байтов для компрессии и записи
//
// Возвращает:
//   - int: количество записанных байтов
//   - error: любые возникающие ошибки при записи
func (g *gzipWriter) Write(data []byte) (int, error) {
	return g.writer.Write(data)
}

// GzipHandle создаёт промежуточную функцию, обеспечивающую сжатие gzip.
// Обрабатывает как сжатие ответов сервера, так и сжатие запросов клиента.
//
// Возвращает:
//   - gin.HandlerFunc: промежуточную функцию, подходящую для использования с Gin
//
// Промежуточная логика:
//   - Сжимает тело запроса, если заголовок Content-Encoding равен "gzip"
//   - Сжимает ответы, если заголовок Accept-Encoding содержит "gzip"
//   - Устанавливает подходящие заголовки Content-Encoding для сжатых ответов
//   - Оборачивает Writer для включения поддержки сжатия
//
// Пример:
//
//	router := gin.Default()
//	router.Use(compress.GzipHandle())
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
				writer:         gz,
			}

			c.Header("Content-Encoding", "gzip")
			c.Header("Content-Type", c.Writer.Header().Get("Content-Type"))
			c.Next()

			return
		}
		c.Next()
	}
}

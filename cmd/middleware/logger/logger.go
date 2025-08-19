package logger

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/sirupsen/logrus"
)

// WithLogging создаёт промежуточную функцию для регистрации HTTP-запросов в Gin.
// Регистрирует подробности каждого HTTP-запроса, включая метод, URI, код состояния,
// длину контента и продолжительность обработки запроса.
//
// Возвращает:
//   - gin.HandlerFunc: промежуточную функцию, совместимую с фреймворком Gin
//
// Промежуточная обработка:
//   - Записывает стартовое время каждого запроса
//   - Пропускает запрос через цепочку промежуточных функций
//   - Вычисляет длительность обработки запроса
//   - Логирует детали запроса с использованием библиотеки logrus
//   - Включает метод, URI, код состояния, длину контента и продолжительность
//
// Поля логирования:
//   - method: HTTP-метод (GET, POST и т.п.)
//   - uri: URI запроса
//   - status_code: Код состояния HTTP-ответа
//   - content_length: Длина содержимого ответа в байтах
//   - duration: Продолжительность обработки запроса
//
// Пример:
//
//	router := gin.Default()
//	router.Use(logger.WithLogging())
func WithLogging() gin.HandlerFunc {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		statusCode := c.Writer.Status()
		contentLength := c.Writer.Size()

		logger.WithFields(logrus.Fields{
			"method":         c.Request.Method,
			"uri":            c.Request.RequestURI,
			"status_code":    statusCode,
			"content_length": contentLength,
			"duration":       duration.String(),
		}).Info("HANDLE REQUEST")
	}
}

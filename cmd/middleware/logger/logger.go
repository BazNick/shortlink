package logger

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/sirupsen/logrus"
)

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
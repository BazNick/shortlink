package functions

import (
	"errors"

	"github.com/gin-gonic/gin"
)

// GetUser - получаем пользователя из запроса.
func GetUser(c *gin.Context) (string, error) {
	uid, ok := c.Get("userID")
	if !ok {
		return "", errors.New("unauthorized")
	}
	userID, ok := uid.(string)
	if !ok || userID == "" {
		return "", errors.New("unauthorized")
	}
	return userID, nil
}

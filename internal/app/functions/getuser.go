package functions

import (
	"errors"

	"github.com/BazNick/shortlink/cmd/middleware/auth"
	"github.com/gin-gonic/gin"
)

func User(c *gin.Context) (string, error) {
	token, err := c.Cookie("token")
	if err != nil {
		return "", err
	}
	userID := auth.GetUserID(token)
	if userID == "" {
		return "", errors.New("unauthorized: invalid user ID")
	}
	return userID, nil
}

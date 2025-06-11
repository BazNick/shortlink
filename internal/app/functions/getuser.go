package functions

import (
	"errors"

	"github.com/BazNick/shortlink/cmd/middleware/auth"
	"github.com/gin-gonic/gin"
)

func User(c *gin.Context, strict bool) (string, error) {
	token, err := c.Cookie("token")
	if err != nil {
		return "", err
	}
	userID := auth.GetUserID(token)
	if userID == "" {
		if strict {
			return "", errors.New("unauthorized: invalid token")
		}
		return "", nil
	}
	return userID, nil
}

package functions

import (
	"net/http"

	"github.com/BazNick/shortlink/cmd/middleware/auth"
	"github.com/gin-gonic/gin"
)

func User(c *gin.Context) string {
	token, err := c.Cookie("token")
	if err != nil {
		c.AbortWithError(http.StatusUnauthorized, err)
	}
	userID := auth.GetUserID(token)
	if userID == "" {
		// c.AbortWithError(http.StatusUnauthorized, err)
	}
	return userID
}

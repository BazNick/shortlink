package auth

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

const (
	CookieName   = "token"
	CookiePath   = "/"
	CookieDomain = ""
	TokenExp     = time.Hour * 3
	SecretKey    = "supersecretkey"
)

func randBytes(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return ``, err
	}
	uuid := fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return uuid, nil
}

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// берем куки
		cookie, err := c.Cookie("token")
		if err == nil && cookie != "" {
			// Кука уже есть — продолжаем цепочку
			c.Next()
			return
		}
		// генерируем последовательность рандомных байт для ID пользователя
		id, err := randBytes(16)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
			},
			UserID: id,
		})

		// создаём строку токена
		tokenString, err := token.SignedString([]byte(SecretKey))
		if err != nil {
			c.AbortWithError(http.StatusUnauthorized, err)
			return
		}

		c.SetCookie(
			CookieName,
			tokenString,
			int(TokenExp.Seconds()),
			CookiePath,
			CookieDomain,
			false,
			true,
		)
		c.Next()
	}
}

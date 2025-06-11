package auth

import (
	"crypto/rand"
	"errors"
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

func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie(CookieName)
		// парсим токен, если есть
		if err == nil && cookie != "" {
			claims, err := ParseToken(cookie, secret)
			if err == nil {
				c.Set("userID", claims.UserID)
				c.Next()
				return
			}
		}

		// Если токена нет — создаём новый
		tokenString, err := GenToken(secret)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		claims, err := ParseToken(tokenString, secret)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
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
		c.Set("userID", claims.UserID)
		c.Next()
	}
}

func GenToken(secretKey string) (string, error) {
	// генерируем последовательность рандомных байт для ID пользователя
	id, err := randBytes(16)
	if err != nil {
		return "", err
	}
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		UserID: id,
	})

	// создаём строку токена
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ParseToken(tokenStr, secret string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(
		tokenStr,
		claims,
		func(token *jwt.Token) (any, error) {
			return []byte(secret), nil
		},
	)
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

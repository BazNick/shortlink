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

// Claims - структура утверждений JWT-токенов.
// Расширяет стандартные утверждения jwt.RegisteredClaims полем UserID.
type Claims struct {
	jwt.RegisteredClaims
	UserID string // UserID - ID пользователя.
}

const (
	CookieName   = "token" // CookieName - название куки
	CookiePath   = "/" // CookiePath - путь до куки
	CookieDomain = "" // CookieDomain - область куки
	TokenExp     = time.Hour * 3 // TokenExp - время жизни токена
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

// Auth создаёт промежуточную функцию для Gin, реализующую аутентификацию на основе JWT.
// Она управляет созданием, проверкой и передачей токенов и идентификации пользователей.
//
// Параметры:
//   - secret: секретный ключ, используемый для подписания и проверки JWT-токенов
//
// Возвращает:
//   - gin.HandlerFunc: промежуточную функцию
//
// Промежуточная логика:
//   - Проверяет наличие действующего токена аутентификации в куках
//   - Валидирует токен, если таковой имеется
//   - Генерирует новый токен, если старый отсутствует или проверка провалилась
//   - Присваивает идентификатор пользователя (userID) для последующих обработчиков
//   - Устанавливает куки с токеном аутентификации в ответ
//
// Пример:
//
//	router := gin.Default()
//	router.Use(auth.Auth("your-secret-key"))
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

// GenToken генерирует новый JWT-токен с уникальным идентификатором пользователя.
//
// Параметры:
//   - secretKey: секретный ключ, используемый для подписания JWT-токена
//
// Возвращает:
//   - string: строковая форма нового JWT-токена
//   - error: возможные ошибки, возникшие при создании токена
//
// Функция:
//   - Генерирует случайный 16-байтовый идентификатор пользователя
//   - Создаёт JWT-токен с подписывающим методом HS256
//   - Устанавливает срок годности токена равным TokenExp (3 часа)
//   - Подписывает токен с помощью переданного секретного ключа
//
// Пример:
//
//	token, err := GenToken("your-secret-key")
//	if err != nil {
//	    log.Fatal(err)
//	}
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

// ParseToken проверяет и парсит JWT-токен.
//
// Параметры:
//   - tokenStr: строка с JWT-токеном
//   - secret: секретный ключ, используемый для проверки токена
//
// Возвращает:
//   - *Claims: разобранные утверждения токена, если он действителен
//   - error: сообщение "invalid token", если токен недействителен или просрочен
//
// Функция:
//   - Анализирует JWT-токен с использованием переданного секретного ключа
//   - Проверяет подпись и срок действия токена
//   - Возвращает утверждения токена, если он действительный
//
// Пример:
//
//	claims, err := ParseToken(tokenString, "secret-key")
//	if err != nil {
//	    log.Printf("Недействительный токен: %v", err)
//	    return
//	}
//	fmt.Printf("Идентификатор пользователя: %s\n", claims.UserID)
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

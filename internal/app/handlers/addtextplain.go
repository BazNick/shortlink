package handlers

import (
	"io"
	"net/http"

	"github.com/BazNick/shortlink/internal/app/apperr"
	"github.com/BazNick/shortlink/internal/app/entities"
	"github.com/BazNick/shortlink/internal/app/functions"
	"github.com/gin-gonic/gin"
)

// AddLink обрабатывает HTTP-запросы типа POST для создания коротких ссылок из обычного текста.
// Принимает обычный текст в теле запроса, содержащий оригинальную ссылку, и возвращает
// укороченную ссылку также в виде простого текста.
//
// Метод ожидает:
//   - метод HTTP POST (возвращает 405 Method Not Allowed для остальных методов)
//   - простой текст тела запроса, содержащий оригинальную ссылку
//   - валидная аутентификация пользователя (извлекается из контекста запроса)
//
// Формат запроса:
//   - Content-Type: text/plain
//   - Body: "https://example.com/very/long/url"
//
// Формат ответа:
//   - Content-Type: text/plain
//   - Body: "https://shortener.ru/abc12345"
//
// Возможные ответы:
//   - 201 Created: если короткая ссылка была успешно создана
//   - 409 Conflict: если ссылка уже существует (возвращает существующую короткую ссылку)
//   - 400 Bad Request: если запрос некорректен или пользователь недействителен
//   - 405 Method Not Allowed: если метод запроса не POST
//   - 500 Internal Server Error: если произошла ошибка операции с хранилищем
//
// Пример:
//
//	POST /
//	Content-Type: text/plain
//	https://example.com/very/long/url
//
//	Ответ: 201 Created
//	Content-Type: text/plain
//	https://shortener.ru/abc12345
func (handler *URLHandler) AddLink(c *gin.Context) {
	user, err := functions.GetUser(c)
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusBadRequest)
		return
	}

	if c.Request.Method != http.MethodPost {
		http.Error(c.Writer, apperr.ErrOnlyPOST.Error(), http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		http.Error(c.Writer, apperr.ErrBodyRead.Error(), http.StatusBadRequest)
		return
	}
	defer c.Request.Body.Close()

	if _, ok := handler.storage.(*entities.DB); !ok {
		alreadyExst := handler.storage.CheckValExists(string(body))
		if alreadyExst {
			http.Error(c.Writer, apperr.ErrLinkExists.Error(), http.StatusBadRequest)
			return
		}
	}

	var (
		randStr  = functions.RandSeq(8)
		hashLink = functions.SchemeAndHost(c.Request) + "/" + randStr
	)

	shortURL, err := handler.storage.AddHash(randStr, string(body), user)
	if err != nil {
		if err == apperr.ErrValAlreadyExists {
			c.Writer.WriteHeader(http.StatusConflict)
			c.Writer.Write([]byte(functions.SchemeAndHost(c.Request) + "/" + shortURL))
			return
		}
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	c.Writer.Header().Set("content-type", "text/plain")
	c.Writer.WriteHeader(http.StatusCreated)
	c.Writer.Write([]byte(hashLink))
}

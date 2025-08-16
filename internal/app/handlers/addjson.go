package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/BazNick/shortlink/internal/app/apperr"
	"github.com/BazNick/shortlink/internal/app/entities"
	"github.com/BazNick/shortlink/internal/app/functions"
	"github.com/gin-gonic/gin"
)

var (
	successResponse = struct {
		Result string `json:"result"`
	}{}

	conflictResponse = struct {
		Result string `json:"result"`
	}{}
)

// PostJSONLink обрабатывает HTTP-запросы типа POST для создания коротких ссылок из JSON-данных.
// Принять JSON-полезную нагрузку с оригинальной ссылкой и вернуть JSON-ответ,
// содержащий сгенерированную короткую ссылку.
//
// Метод ожидает:
//   - метод HTTP POST (возвращает 405 Method Not Allowed для остальных методов)
//   - тело запроса в формате JSON с полем "url", содержащим оригинальную ссылку
//   - валидная аутентификация пользователя (извлекается из контекста запроса)
//
// Формат запроса:
//
//	{
//	  "url": "https://example.com/very/long/url"
//	}
//
// Формат ответа:
//
//	{
//	  "result": "https://shortener.ru/abc12345"
//	}
//
// Возможные ответы:
//   - 201 Created: если короткая ссылка успешно создана
//   - 409 Conflict: если ссылка уже существует (возвращает существующую короткую ссылку)
//   - 400 Bad Request: если запрос некорректен или пользователь недействителен
//   - 405 Method Not Allowed: если метод запроса не POST
//   - 500 Internal Server Error: если операция с хранилищем завершилась неудачей
//
// Пример:
//
//	POST /api/shorten
//	Content-Type: application/json
//	{
//	  "url": "https://example.com/very/long/url"
//	}
//
//	Ответ: 201 Created
//	{
//	  "result": "https://shortener.ru/abc12345"
//	}
func (handler *URLHandler) PostJSONLink(c *gin.Context) {
    user, err := functions.GetUser(c)
    if err != nil {
        http.Error(c.Writer, err.Error(), http.StatusBadRequest)
        return
    }

    if c.Request.Method != http.MethodPost {
        http.Error(c.Writer, apperr.ErrOnlyPOST.Error(), http.StatusMethodNotAllowed)
        return
    }

    var link JSONLink
    if err := json.NewDecoder(c.Request.Body).Decode(&link); err != nil {
        http.Error(c.Writer, err.Error(), http.StatusBadRequest)
        return
    }

    if _, ok := handler.storage.(*entities.DB); !ok {
        alreadyExst := handler.storage.CheckValExists(link.Link)
        if alreadyExst {
            http.Error(c.Writer, apperr.ErrLinkExists.Error(), http.StatusBadRequest)
            return
        }
    }

    baseURL := functions.SchemeAndHost(c.Request)

    var (
        randStr  = functions.RandSeq(8) // генерируем случайную последовательность длиной 8 символов
        hashLink = baseURL + "/" + randStr // формируем полную короткую ссылку
    )

    shortURL, err := handler.storage.AddHash(randStr, link.Link, user)
    if err != nil {
        if err.Error() == "conflict" {
            conflictResponse.Result = baseURL + "/" + shortURL
            resp, err := json.Marshal(conflictResponse)
            if err != nil {
                http.Error(c.Writer, err.Error(), http.StatusBadRequest)
                return
            }

            c.Writer.Header().Set("content-type", "application/json")
            c.Writer.WriteHeader(http.StatusConflict)
            c.Writer.Write(resp)
            return
        }
        http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
        return
    }

    successResponse.Result = hashLink
    resp, err := json.Marshal(successResponse)
    if err != nil {
        http.Error(c.Writer, err.Error(), http.StatusBadRequest)
        return
    }

    c.Writer.Header().Set("content-type", "application/json")
    c.Writer.WriteHeader(http.StatusCreated)
    c.Writer.Write(resp)
}

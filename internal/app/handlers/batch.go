package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/BazNick/shortlink/internal/app/apperr"
	"github.com/BazNick/shortlink/internal/app/entities"
	"github.com/BazNick/shortlink/internal/app/functions"
	"github.com/gin-gonic/gin"
)

// BatchLinks обрабатывает HTTP-запросы POST для создания нескольких коротких URL в рамках одного пакетного запроса.
// Принимает массив JSON с исходными URL и возвращает массив с короткими URL и соответствующими идентификационными номерами запросов.
//
// Ожидает:
//   - метод HTTP POST (возвращает 405 Method Not Allowed для прочих методов)
//   - тело запроса в формате JSON, содержащее массив объектов BatchIn
//   - валидную аутентификацию пользователя (получаемую из контекста запроса)
//
// Формат запроса:
//
//	[
//	  {
//	    "correlation_id": "req-1",
//	    "original_url": "https://example1.com"
//	  },
//	  {
//	    "correlation_id": "req-2",
//	    "original_url": "https://example2.com"
//	  }
//	]
//
// Формат ответа:
//
//	[
//	  {
//	    "correlation_id": "req-1",
//	    "short_url": "https://shortener.com/abc123"
//	  },
//	  {
//	    "correlation_id": "req-2",
//	    "short_url": "https://shortener.com/def456"
//	  }
//	]
//
// Возможные ответы:
//   - 201 Created: если все короткие URL были успешно созданы
//   - 400 Bad Request: если запрос неверен, пользователь невалиден или какая-то из URL уже существует
//   - 405 Method Not Allowed: если метод запроса не POST
//   - 500 Internal Server Error: если произошли ошибки в работе с БД
//
// Функционал:
//   - Проверяет, что ни одна из URL пакета заранее не существует
//   - Для хранения в БД применяет транзакции, обеспечивая атомарность операций
//
// Пример:
//
//	POST /api/shorten/batch
//	Content-Type: application/json
//	[
//	  {
//	    "correlation_id": "req-1",
//	    "original_url": "https://example1.com"
//	  }
//	]
//
//	Ответ: 201 Created
//	Content-Type: application/json
//	[
//	  {
//	    "correlation_id": "req-1",
//	    "short_url": "https://shortener.com/abc123"
//	  }
//	]
func (handler *URLHandler) BatchLinks(c *gin.Context) {
	user, err := functions.GetUser(c)
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusBadRequest)
		return
	}

	if c.Request.Method != http.MethodPost {
		http.Error(c.Writer, apperr.ErrOnlyPOST.Error(), http.StatusMethodNotAllowed)
		return
	}

	var links []BatchIn

	if err := json.NewDecoder(c.Request.Body).Decode(&links); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusBadRequest)
		return
	}

	baseURL := functions.SchemeAndHost(c.Request)

	for _, link := range links {
		if handler.storage.CheckValExists(link.OriginalURL) {
			http.Error(c.Writer, apperr.ErrLinkExists.Error(), http.StatusBadRequest)
			return
		}
	}

	out := make([]BatchOut, len(links))

	// если это БД
	if _, ok := handler.storage.(*entities.DB); ok {
		tx, err := handler.db.Begin()
		if err != nil {
			http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
			return
		}

		for idx, link := range links {
			shortURL := functions.RandSeq(8)
			_, err := tx.ExecContext(
				context.Background(),
				`INSERT INTO links (short_url, original_url, user_id) VALUES ($1, $2, $3)`,
				shortURL,
				link.OriginalURL,
				user,
			)
			if err != nil {
				tx.Rollback()
				http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
				return
			}
			out[idx].CorrelationID = link.CorrelationID
			out[idx].ShortURL = baseURL + "/" + shortURL
		}

		if err := tx.Commit(); err != nil {
			http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// если это не БД, то сохраняем в файл или в мапу
		for idx, link := range links {
			shortURL := functions.RandSeq(8)
			handler.storage.AddHash(shortURL, link.OriginalURL, user)
			out[idx].CorrelationID = link.CorrelationID
			out[idx].ShortURL = baseURL + "/" + shortURL
		}
	}

	resp, err := json.Marshal(out)

	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusBadRequest)
		return
	}

	c.Writer.Header().Set("content-type", "application/json")
	c.Writer.WriteHeader(http.StatusCreated)
	c.Writer.Write(resp)
}

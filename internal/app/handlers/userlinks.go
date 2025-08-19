package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/BazNick/shortlink/internal/app/entities"
	"github.com/BazNick/shortlink/internal/app/functions"
	"github.com/gin-gonic/gin"
)

// GetUserLinks обрабатывает HTTP-запросы GET для получения всех коротких URL, созданных конкретным пользователем.
// Запрашивает БД для получения всех URL, связанных с авторизованным пользователем,
// и возвращает их в виде массива JSON.
//
// Метод ожидает:
//   - метод HTTP GET
//   - валидную аутентификацию пользователя (получаемую из контекста запроса)
//   - тело запроса не требуется
//
// Формат ответа:
//
//	[
//	  {
//	    "short_url": "https://shortener.com/abc123",
//	    "original_url": "https://example1.com"
//	  },
//	  {
//	    "short_url": "https://shortener.com/def456",
//	    "original_url": "https://example2.com"
//	  }
//	]
//
// Возможные ответы:
//   - 200 OK: если найдены и успешно возвращены URL
//   - 204 No Content: если для пользователя не было найдено никаких URL
//   - 401 Unauthorized: если аутентификация пользователя не удалась
//   - 400 Bad Request: если возникла ошибка при обработке результатов
//   - 500 Internal Server Error: если операции с базой данных завершились неудачно
//
// Функционал:
//   - Проверяет валидность аутентификации пользователя
//   - Запрашивает БД для получения всех URL, связанных с данным пользователем
//   - Формирует полные короткие URL, объединяя базовый URL и хэш короткого URL
//   - Возвращает результаты в виде массива JSON
//   - Возвращает 204 No Content, если не найдено ни одной записи
//
// Пример:
//
//	GET /api/user/urls
//
//	Ответ: 200 OK
//	Content-Type: application/json
//	[
//	  {
//	    "short_url": "https://shortener.com/abc123",
//	    "original_url": "https://example1.com"
//	  }
//	]
//
//	Ответ: 204 No Content (если у пользователя нет URL)
func (handler *URLHandler) GetUserLinks(c *gin.Context) {
	user, err := functions.GetUser(c)
	if err != nil {
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	rows, err := handler.db.QueryContext(
		context.Background(),
		`SELECT short_url, original_url FROM links WHERE user_id = $1`,
		user,
	)
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	baseURL := functions.SchemeAndHost(c.Request)
	result := make([]entities.FileLinks, 0, 10)

	for rows.Next() {
		var rec entities.FileLinks
		if err := rows.Scan(&rec.ShortURL, &rec.OriginalURL); err != nil {
			http.Error(c.Writer, err.Error(), http.StatusBadRequest)
			return
		}
		rec.ShortURL = baseURL + "/" + rec.ShortURL
		result = append(result, rec)
	}

	if len(result) == 0 {
		c.Writer.WriteHeader(http.StatusNoContent)
		return
	}

	if err := rows.Err(); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(result)

	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusBadRequest)
		return
	}

	c.Writer.Header().Set("content-type", "application/json")
	c.Writer.Write(resp)
}

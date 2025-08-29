package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/BazNick/shortlink/internal/app/entities"
	"github.com/BazNick/shortlink/internal/app/functions"
	"github.com/gin-gonic/gin"
)

// DeleteUserLinks обрабатывает HTTP-запросы DELETE для пометки множества коротких URL как удалённых.
// Принимает массив JSON с хэшами коротких URL и обрабатывает операцию удаления асинхронно.
//
// Метод ожидает:
//   - метод HTTP DELETE
//   - тело запроса в формате JSON, содержащее массив хэшей коротких URL
//   - валидную аутентификацию пользователя (получаемую из контекста запроса)
//
// Формат запроса:
//
//	["abc123", "def456", "ghi789"]
//
// Возможные ответы:
//   - 202 Accepted: если запрос на удаление принят для дальнейшей обработки
//   - 401 Unauthorized: если аутентификация пользователя не прошла
//   - 400 Bad Request: если структура запроса неправильная
//
// Функционал:
//   - Проверяет валидность аутентификации пользователя
//   - Парсит массив JSON с хэшами коротких URL
//   - Передаёт запрос DeleteRequest в канал DeleteChan для асинхронной обработки
//   - Немедленно возвращает статус 202 Accepted
//   - Само удаление выполняется фоновыми рабочими горутинами, слушающими канал DeleteChan
//
// Пример:
//
//	DELETE /api/user/urls
//	Content-Type: application/json
//	["abc123", "def456"]
//
//	Ответ: 202 Accepted
func (handler *URLHandler) DeleteUserLinks(c *gin.Context) {
	user, err := functions.GetUser(c)
	if err != nil {
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	var links []string
	if err := json.NewDecoder(c.Request.Body).Decode(&links); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusBadRequest)
		return
	}

	if handler.workerManager != nil {
		handler.workerManager.SendDeleteRequest(entities.DeleteRequest{
			UserID:    user,
			ShortURLs: links,
		})
	}

	c.Writer.WriteHeader(http.StatusAccepted)
}

package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/BazNick/shortlink/internal/app/service"
	"github.com/gin-gonic/gin"
)

// StatsResponse представляет ответ API для статистики сервиса.
type StatsResponse struct {
	URLs  int `json:"urls"`  // Количество сокращённых URL в сервисе
	Users int `json:"users"` // Количество пользователей в сервисе
}

// GetStats обрабатывает HTTP-запросы типа GET для получения статистики сервиса.
// Возвращает JSON-ответ с количеством сокращённых URL и пользователей.
//
// Метод ожидает:
//   - метод HTTP GET (возвращает ошибку 405 Method Not Allowed для остальных методов)
//   - валидная аутентификация пользователя (извлекается из контекста запроса)
//
// Формат ответа:
//
//	{
//	  "urls": 150,
//	  "users": 25
//	}
//
// Возможные ответы:
//   - 200 OK: если статистика успешно получена
//   - 405 Method Not Allowed: если метод запроса не GET
//   - 500 Internal Server Error: если операция с хранилищем завершилась неудачей
//
// Пример:
//
//	GET /api/internal/stats
//
//	Ответ: 200 OK
//	{
//	  "urls": 150,
//	  "users": 25
//	}
func (handler *URLHandler) GetStats(c *gin.Context) {
	if c.Request.Method != http.MethodGet {
		http.Error(c.Writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Используем сервис для получения статистики
	req := service.GetStatsRequest{}
	resp, err := handler.urlService.GetStats(c.Request.Context(), req)
	if err != nil {
		http.Error(c.Writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := StatsResponse{
		URLs:  resp.URLs,
		Users: resp.Users,
	}

	jsonResp, err := json.Marshal(response)
	if err != nil {
		http.Error(c.Writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Write(jsonResp)
}

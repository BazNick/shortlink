package handlers

import (
	"net/http"

	"github.com/BazNick/shortlink/internal/app/apperr"
	"github.com/BazNick/shortlink/internal/app/service"
	"github.com/gin-gonic/gin"
)

// GetLink обрабатывает HTTP-запросы типа GET для получения и переадресации на оригинальные ссылки.
// Извлекает короткий хэш URL из параметра URL, ищет соответствующую оригинальную ссылку
// и производит временную переадресацию 307 на оригинальный URL.
//
// Метод ожидает:
//   - метод HTTP GET (возвращает ошибку 405 Method Not Allowed для остальных методов)
//   - параметр URL "id", содержащий короткий хэш URL
//
// Ответы:
//   - 307 Temporary Redirect: если короткий URL найден, переадресует на оригинал
//   - 410 Gone: если короткий URL не найден
//   - 405 Method Not Allowed: если метод запроса не GET
//
// Пример:
//
//	GET /abc123
//	Ответ: 307 Temporary Redirect с заголовком Location, установленным на оригинальный URL
func (handler *URLHandler) GetLink(c *gin.Context) {
	if c.Request.Method != http.MethodGet {
		http.Error(c.Writer, apperr.ErrOnlyGET.Error(), http.StatusMethodNotAllowed)
		return
	}

	id := c.Param("id") // извлекаем id короткого URL из маршрута

	// Используем сервис для получения оригинальной ссылки
	req := service.GetOriginalURLRequest{ShortURL: id}
	resp, err := handler.urlService.GetOriginalURL(c.Request.Context(), req)
	if err != nil {
		http.Error(c.Writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !resp.Found { // если оригинальная ссылка не найдена
		http.Error(c.Writer, apperr.ErrLinkNotFound.Error(), http.StatusGone)
		return
	}

	// устанавливаем заголовки для временной переадресации
	c.Writer.Header().Set("Location", resp.OriginalURL)
	c.Writer.Header().Set("Content-Type", "text/html")
	c.Writer.WriteHeader(http.StatusTemporaryRedirect)
}

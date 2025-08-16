package handlers

import (
	"database/sql"

	"github.com/BazNick/shortlink/internal/app/entities"
	"github.com/BazNick/shortlink/internal/app/storage"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type (
	// JSONLink представляет структуру JSON-запросов для сокращения ссылок.
	// Содержит оригинальную ссылку, которую нужно сократить.
	JSONLink struct {
		Link string `json:"url"` // Оригинальная ссылка для сокращения
	}

	// URLHandler обрабатывает HTTP-запросы для операций сокращения ссылок.
	// Предоставляет методы для создания коротких ссылок и перенаправления на оригинальные.
	URLHandler struct {
		storage storage.Storage // Интерфейс хранения для сопоставления ссылок
		path    string          // Путь к файлу для файлового хранилища
		dbPath  string          // Строка подключения базы данных
		db      *sql.DB         // Подключение к базе данных (если используется хранение в БД)
	}

	// BatchIn представляет одну ссылку в множественном запросе на сокращение.
	// Включает уникальный идентификатор для связи ответа с запросом.
	BatchIn struct {
		CorrelationID string `json:"correlation_id"` // Уникальный идентификатор этой ссылки в пакете
		OriginalURL   string `json:"original_url"`   // Оригинальная ссылка для сокращения
	}

	// BatchOut представляет ответ для одной ссылки в пакетном запросе на сокращение.
	// Включает уникальный идентификатор и сгенерированную короткую ссылку.
	BatchOut struct {
		CorrelationID string `json:"correlation_id"` // Соответствует уникальный идентификатору из запроса
		ShortURL      string `json:"short_url"`      // Сгенерированная короткая ссылка
	}
)

// NewURLHandler создаёт и возвращает новый экземпляр URLHandler.
// Инициализирует обработчик с предоставленным хранилищем и конфигурацией.
//
// Параметры:
//   - storage: реализация хранилища
//   - filePath: путь к файлу для файлового хранилища
//   - dbPath: строка соединения базы данных для хранения в БД
//
// Возвращает:
//   - *URLHandler: новый экземпляр URLHandler
//
// Пример:
//
//	storage := entities.NewHashDict()
//	handler := NewURLHandler(storage, "urls.json", "postgres://user:pass@localhost/db")
func NewURLHandler(
	storage storage.Storage,
	filePath, dbPath string,
) *URLHandler {
	var db *sql.DB

	if dbStorage, ok := storage.(*entities.DB); ok {
		db = dbStorage.Database
	}

	handler := &URLHandler{
		storage: storage,
		path:    filePath,
		dbPath:  dbPath,
		db:      db,
	}

	return handler
}

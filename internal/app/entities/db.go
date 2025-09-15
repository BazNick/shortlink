package entities

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/BazNick/shortlink/internal/app/apperr"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// DB - соединение с БД, реализует интерфейс Storage.
type DB struct {
	Database *sql.DB
}

// NewDB создаёт новое подключение к БД и инициализирует необходимые таблицы.
// Устанавливает соединение с БД и создаёт таблицу ссылок, если её ещё нет.
//
// Параметры:
//   - connection: строка подключения к PostgreSQL (например, "postgres://user:password@localhost:5432/dbname")
//
// Возвращает:
//   - *DB: новый экземпляр DB с инициализированным подключением к БД
//   - nil: если подключение или создание таблиц завершилось ошибкой
//
// Функционал:
//   - Открывает соединение с PostgreSQL с использованием переданной строки подключения
//   - Выполняет проверку доступности БД
//   - Создаёт таблицу ссылок с подходящей схемой, если таблицы ещё не существует
//   - Создаёт индекс на поле original_url для эффективного поиска
//
// Схема таблицы:
//   - short_url: varchar(15) NOT NULL, PRIMARY KEY
//   - original_url: text NOT NULL UNIQUE
//   - user_id: text NOT NULL
//   - is_deleted: BOOLEAN DEFAULT FALSE
//
// Пример:
//
//	db := NewDB("postgres://user:password@localhost:5432/shortlink")
//	if db == nil {
//	    log.Fatal("Ошибка подключения к базе данных")
//	}
func NewDB(connection string) *DB {
	db, err := sql.Open("pgx", connection)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		fmt.Println(err)
		return nil
	}

	_, err = db.ExecContext(
		ctx,
		`CREATE TABLE IF NOT EXISTS links (
			short_url varchar(15) NOT NULL,
			original_url text NOT NULL UNIQUE,
			user_id text NOT NULL,
			is_deleted BOOLEAN DEFAULT FALSE,
			PRIMARY KEY (short_url)
		)`,
	)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	_, err = db.ExecContext(
		ctx,
		`CREATE INDEX IF NOT EXISTS idx_original_url ON links(original_url);`,
	)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return &DB{Database: db}
}

// AddHash сохраняет новое сопоставление хэша с ссылкой в БД.
// Реализует метод AddHash интерфейса Storage.
//
// Параметры:
//   - hash: короткий хэш URL для сохранения
//   - link: оригинальная ссылка, связанная с хэшем
//   - userID: идентификатор пользователя
//
// Возвращает:
//   - string: сохранённый короткий хэш URL
//   - error: ErrValAlreadyExists, если ссылка уже существует, либо другие ошибки базы данных
//
// Функционал:
//   - Вставляет новую запись в таблицу links
//   - Обрабатывает конфликты уникальных ключей путём возврата существующей короткой ссылки
//   - Возвращает ErrValAlreadyExists, если ссылка уже существует
//
// Пример:
//
//	shortURL, err := db.AddHash("abc123", "https://example.com", "user123")
//	if err == apperr.ErrValAlreadyExists {
//	    fmt.Printf("Ссылка уже существует с коротким URL: %s\n", shortURL)
//	}
func (db *DB) AddHash(hash, link, userID string) (string, error) {
	var shortURL string

	err := db.Database.QueryRowContext(
		context.Background(),
		`INSERT INTO links (short_url, original_url, user_id) 
		 VALUES ($1, $2, $3) 
		 RETURNING short_url;`,
		hash,
		link,
		userID,
	).Scan(&shortURL)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			errQueryRow := db.Database.QueryRowContext(
				context.Background(),
				`SELECT short_url FROM links WHERE original_url = $1`,
				link,
			).Scan(&shortURL)

			if errQueryRow != nil {
				return "", fmt.Errorf("conflict, but failed to retrieve short_url: %w", errQueryRow)
			}

			return shortURL, apperr.ErrValAlreadyExists
		}

		return "", err
	}

	return shortURL, nil
}

// GetHash получает оригинальную ссылку, ассоциированную с заданным хэшем.
// Реализует метод GetHash интерфейса Storage.
//
// Параметры:
//   - hash: короткий хэш URL для поиска
//
// Возвращает:
//   - string: оригинальную ссылку, если она найдена и не удалена, пустую строку в другом случае
//
// Функционал:
//   - Запрашивает БД для нахождения оригинальной ссылки по данному хэшу
//   - Возвращает только те ссылки, которые не помечены как удалённые (is_deleted = false)
//   - Возвращает пустую строку, если хэш не найден или ссылка удалена
//
// Пример:
//
//	originalURL := db.GetHash("abc123")
//	if originalURL == "" {
//	    fmt.Println("Ссылка не найдена или была удалена")
//	}
func (db *DB) GetHash(hash string) string {
	row := db.Database.QueryRowContext(
		context.Background(),
		`SELECT original_url FROM links WHERE short_url = $1 AND is_deleted = false;`,
		hash,
	)

	var link string

	err := row.Scan(&link)
	if err != nil {
		return ""
	}

	return link
}

// CheckValExists проверяет, существует ли оригинальная ссылка в БД.
// Реализует метод CheckValExists интерфейса Storage.
//
// Параметры:
//   - link: оригинальная ссылка для проверки существования
//
// Возвращает:
//   - bool: true, если ссылка существует в БД, false в обратном случае
//
// Функционал:
//   - Запрашивает БД для проверки наличия оригинальной ссылки
//   - Использует sql.NullString для обработки случаев отсутствия записей
//   - Возвращает true, если ссылка существует, false в противном случае
//
// Пример:
//
//	exists := db.CheckValExists("https://example.com")
//	if exists {
//	    fmt.Println("Ссылка уже существует в базе данных")
//	}
func (db *DB) CheckValExists(link string) bool {
	row := db.Database.QueryRowContext(
		context.Background(),
		`SELECT original_url FROM links WHERE original_url = $1`,
		link,
	)

	var exists sql.NullString

	err := row.Scan(&exists)

	if err != nil {
		fmt.Println(err)
	}

	if exists.Valid {
		return true
	}

	return false
}

// GetStats возвращает статистику хранилища DB.
//
// Возвращает:
//   - urls: количество сокращённых URL в сервисе
//   - users: количество пользователей в сервисе
//   - error: ошибка базы данных, если произошла
//
// Функционал:
//   - Выполняет SQL-запросы для подсчёта количества URL и уникальных пользователей
//   - Использует COUNT(DISTINCT user_id) для подсчёта уникальных пользователей
//   - Возвращает статистику или ошибку базы данных
//
// Пример:
//
//	urls, users, err := db.GetStats()
//	if err != nil {
//	    log.Printf("Ошибка получения статистики: %v", err)
//	} else {
//	    fmt.Printf("URLs: %d, Users: %d\n", urls, users)
//	}
func (db *DB) GetStats() (int, int, error) {
	ctx := context.Background()

	var urls int
	err := db.Database.QueryRowContext(ctx, `SELECT COUNT(*) FROM links`).Scan(&urls)
	if err != nil {
		return 0, 0, err
	}

	var users int
	err = db.Database.QueryRowContext(ctx, `SELECT COUNT(DISTINCT user_id) FROM links`).Scan(&users)
	if err != nil {
		return 0, 0, err
	}

	return urls, users, nil
}

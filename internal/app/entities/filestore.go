package entities

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"
)

// FileLinks - хранение данных в файле.
type FileLinks struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id"`
}

// FileStore - реализацию хранилища на основе файла для сопоставления URL.
type FileStore struct {
	Path        string
	FileStorage *os.File
	mu          sync.RWMutex
	cache       map[string]string
	users       map[string]bool
}

// NewFileStore создаёт новый экземпляр хранилища на основе файла.
// Инициализирует кэш и загружает существующие данные из файла.
//
// Параметры:
//   - path: путь к файлу, где будут храниться URL
//
// Возвращает:
//   - *FileStore: новый экземпляр FileStore с инициализированным кэшем
//
// Функция:
//   - Создаёт новый FileStore с указанным путём к файлу
//   - Инициализирует пустой кэш
//   - Загружает имеющиеся URL из файла в кэш
//
// Пример:
//
//	store := NewFileStore("/path/to/urls.json")
//	shortURL, err := store.AddHash("abc123", "https://example.com", "user123")
func NewFileStore(path string) *FileStore {
	fs := &FileStore{
		Path:  path,
		cache: make(map[string]string),
		users: make(map[string]bool),
	}

	fs.loadCache()
	return fs
}

// loadCache загружает URL из файла хранилища в память.
// Читает файл построчно, разбирая каждую строку как объект JSON формата FileLinks.
// Если файл не существует, создаётся пустой файл.
func (f *FileStore) loadCache() {
	file, err := os.OpenFile(f.Path, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var res FileLinks
		if err := json.Unmarshal([]byte(scanner.Text()), &res); err == nil {
			f.cache[res.ShortURL] = res.OriginalURL
			if res.UserID != "" {
				f.users[res.UserID] = true
			}
		}
	}
}

// AddHash сохраняет ссылки в файловое хранилище.
// Реализует метод AddHash интерфейса Storage.
//
// Параметры:
//   - hash: короткий URL-хэш для сохранения
//   - link: исходный URL, связанный с хэшем
//   - userID: идентификатор пользователя
//
// Возвращает:
//   - string: короткий URL-хэш, который был сохранён
//   - error: любую ошибку, произошедшую при хранении
//
// Функция:
//   - Проверяет, существует ли хэш в кэше
//   - Добавляет URL в кэш
//   - Дополняет файл хранилища записью в формате JSON
//   - Применяет потокобезопасные операции с блокировкой
//
// Пример:
//
//	shortURL, err := store.AddHash("abc123", "https://example.com", "user123")
//	if err != nil {
//	    log.Printf("Не удалось сохранить URL: %v", err)
//	}
func (f *FileStore) AddHash(hash, link, userID string) (string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if _, exists := f.cache[hash]; exists {
		return hash, nil
	}

	f.cache[hash] = link
	if userID != "" {
		f.users[userID] = true
	}

	file, err := os.OpenFile(f.Path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return "", err
	}
	defer file.Close()

	data, err := json.Marshal(FileLinks{
		ShortURL:    hash,
		OriginalURL: link,
		UserID:      userID,
	})
	if err != nil {
		return "", err
	}

	writer := bufio.NewWriter(file)
	writer.Write(data)
	writer.WriteRune('\n')
	writer.Flush()

	return "", nil
}

// GetHash получает исходный URL.
// Реализует метод GetHash интерфейса Storage.
//
// Параметры:
//   - hash: короткий URL-хэш для поиска
//
// Возвращает:
//   - string: исходный URL, если найден, пустую строку в противном случае
//
// Функция:
//   - Производит поиск хэша в кэше
//   - Возвращает соответствующий исходный URL, если найдено совпадение
//   - Возвращает пустую строку, если хэш не найден
//   - Применяет потокобезопасные операции с блокировкой
//
// Пример:
//
//	originalURL := store.GetHash("abc123")
//	if originalURL == "" {
//	    fmt.Println("URL не найден")
//	}
func (f *FileStore) GetHash(hash string) string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if link, exists := f.cache[hash]; exists {
		return link
	}

	return ""
}

// CheckValExists проверяет, существует ли исходный URL в хранилище.
// Реализует метод CheckValExists интерфейса Storage.
//
// Параметры:
//   - link: исходный URL для проверки на существование
//
// Возвращает:
//   - bool: true, если URL существует в хранилище, false в противном случае
//
// Функция:
//   - Перебирает все закэшированные пары URL
//   - Возвращает true, если исходный URL обнаружен
//   - Возвращает false, если URL не найден
//   - Применяет потокобезопасные операции с блокировкой
//
// Пример:
//
//	exists := store.CheckValExists("https://example.com")
//	if exists {
//	    fmt.Println("URL уже существует в хранилище")
//	}
func (f *FileStore) CheckValExists(link string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	for _, cachedLink := range f.cache {
		if cachedLink == link {
			return true
		}
	}

	return false
}

// GetStats возвращает статистику хранилища FileStore.
//
// Возвращает:
//   - urls: количество сокращённых URL в сервисе
//   - users: количество пользователей в сервисе
//   - error: всегда nil для FileStore
//
// Пример:
//
//	urls, users, err := store.GetStats()
//	fmt.Printf("URLs: %d, Users: %d\n", urls, users)
func (f *FileStore) GetStats() (int, int, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	urls := len(f.cache)
	users := len(f.users)
	return urls, users, nil
}

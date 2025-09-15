package entities

// HashDict представляет отображение между короткими URL и оригинальными URL.
// Поддерживает две мапы: одна для отображения хэша на ссылку, другая для обратного отображения ссылки на хэш.
type HashDict struct {
	Dict    map[string]string
	RevDict map[string]string
	Users   map[string]bool
}

// NewHashDict создает и возвращает новый экземпляр HashDict с инициализированными мапами.
//
// Возвращает:
//   - *HashDict: новый экземпляр HashDict
//
// Пример:
//
//	hd := NewHashDict()
//	hd.AddHash("abc123", "https://example.com", "user1")
func NewHashDict() *HashDict {
	return &HashDict{
		Dict:    make(map[string]string),
		RevDict: make(map[string]string),
		Users:   make(map[string]bool),
	}
}

// AddHash добавляет новое отображение хэша на ссылку в мапу.
// Также создаёт обратное отображение от ссылки к хэшу для быстрого поиска.
//
// Параметры:
//   - hash: короткий хэш URL
//   - link: оригинальная ссылка
//   - userID: идентификатор пользователя
//
// Возвращает:
//   - string: возвращает пустую строку
//   - error: возвращает nil
//
// Пример:
//
//	hd := NewHashDict()
//	_, err := hd.AddHash("abc123", "https://example.com", "user1")
func (hasdDict *HashDict) AddHash(hash, link, userID string) (string, error) {
	hasdDict.Dict[hash] = link
	hasdDict.RevDict[link] = hash
	if userID != "" {
		hasdDict.Users[userID] = true
	}
	return "", nil
}

// GetHash получает оригинальную ссылку, ассоциированную с указанным хэшем.
//
// Параметры:
//   - hash: короткий хэш URL для поиска
//
// Возвращает:
//   - string: оригинальная ссылка, если найдена, пустая строка в ином случае
//
// Пример:
//
//	originalURL := hd.GetHash("abc123") // Вернёт "https://example.com"
func (hasdDict *HashDict) GetHash(hash string) string {
	return hasdDict.Dict[hash]
}

// CheckValExists проверяет, существует ли данная оригинальная ссылка в мапе.
//
// Параметры:
//   - link: оригинальная ссылка для проверки
//
// Возвращает:
//   - bool: true, если ссылка существует, false в ином случае
//
// Пример:
//
//	exists := hd.CheckValExists("https://example.com")
func (hasdDict *HashDict) CheckValExists(link string) bool {
	_, exists := hasdDict.RevDict[link]
	return exists
}

// GetStats возвращает статистику хранилища HashDict.
//
// Возвращает:
//   - urls: количество сокращённых URL в сервисе
//   - users: количество пользователей в сервисе
//   - error: всегда nil для HashDict
//
// Пример:
//
//	urls, users, err := hd.GetStats()
//	fmt.Printf("URLs: %d, Users: %d\n", urls, users)
func (hasdDict *HashDict) GetStats() (int, int, error) {
	urls := len(hasdDict.Dict)
	users := len(hasdDict.Users)
	return urls, users, nil
}

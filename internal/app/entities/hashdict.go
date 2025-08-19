package entities

// HashDict представляет отображение между короткими URL и оригинальными URL.
// Поддерживает две мапы: одна для отображения хэша на ссылку, другая для обратного отображения ссылки на хэш.
type HashDict struct {
	Dict    map[string]string
	RevDict map[string]string
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
	}
}

// AddHash добавляет новое отображение хэша на ссылку в мапу.
// Также создаёт обратное отображение от ссылки к хэшу для быстрого поиска.
//
// Параметры:
//   - hash: короткий хэш URL
//   - link: оригинальная ссылка
//   - _: идентификатор пользователя (здесь не используется)
//
// Возвращает:
//   - string: возвращает пустую строку
//   - error: возвращает nil
//
// Пример:
//
//	hd := NewHashDict()
//	_, err := hd.AddHash("abc123", "https://example.com", "")
func (hasdDict *HashDict) AddHash(hash, link, _ string) (string, error) {
	hasdDict.Dict[hash] = link
	hasdDict.RevDict[link] = hash
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

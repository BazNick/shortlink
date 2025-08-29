package entities

import "database/sql"

// DeleteRequest - удаление нескольких коротких URL конкретного пользователя.
type DeleteRequest struct {
	UserID    string   // ID пользователя.
	ShortURLs []string // ShortURLs - слайс коротких ссылок
}

// DeleteWorkerManager управляет горутинами для асинхронной обработки запросов на удаление.
type DeleteWorkerManager struct {
	deleteChan chan DeleteRequest
	db         *sql.DB
}

// NewDeleteWorkerManager создает новый экземпляр DeleteWorkerManager.
func NewDeleteWorkerManager(db *sql.DB, bufferSize int) *DeleteWorkerManager {
	return &DeleteWorkerManager{
		deleteChan: make(chan DeleteRequest, bufferSize),
		db:         db,
	}
}

// StartDeleteWorkers запускает указанное количество горутин для обработки запросов на удаление.
// Каждая горутина прослушивает канал и помечает URL как удалённые в БД.
//
// Параметры:
//   - workerCount: число запускаемых горутин
//
// Функция:
//   - Запускает указанное количество горутин
//   - Каждая горутина слушает канал на предмет поступающих запросов на удаление
//   - Горутины выполняют SQL-команды обновления для пометки URL как удалённые
//   - Генерирует panic, если возникают ошибки в операциях с базой данных
//
// Горутины:
//   - Обрабатывают объекты DeleteRequest из канала
//   - Обновляют БД, помечая URL как удалённые для указанного пользователя
//
// Пример:
//
//	manager := NewDeleteWorkerManager(db, 100)
//	manager.StartDeleteWorkers(5) // Запустить 5 горутин
//
//	// Отправить запрос на удаление
//	manager.SendDeleteRequest(DeleteRequest{
//	    UserID:    "user123",
//	    ShortURLs: []string{"abc123", "def456"},
//	})
func (dwm *DeleteWorkerManager) StartDeleteWorkers(workerCount int) {
	for i := 0; i < workerCount; i++ {
		go func(id int) {
			for req := range dwm.deleteChan {
				_, err := dwm.db.Exec(
					`UPDATE links SET is_deleted = true WHERE user_id = $1 AND short_url = ANY($2);`,
					req.UserID,
					req.ShortURLs,
				)
				if err != nil {
					panic(err)
				}
			}
		}(i)
	}
}

// SendDeleteRequest отправляет запрос на удаление в канал для обработки.
func (dwm *DeleteWorkerManager) SendDeleteRequest(req DeleteRequest) {
	dwm.deleteChan <- req
}

// Close закрывает канал и останавливает все горутины.
func (dwm *DeleteWorkerManager) Close() {
	close(dwm.deleteChan)
}

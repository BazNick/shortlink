package entities

import "database/sql"

// DeleteRequest - удаление нескольких коротких URL конкретного пользователя.
type DeleteRequest struct {
	UserID    string // ID пользователя.
	ShortURLs []string // ShortURLs - слайс коротких ссылок
}

// DeleteChan - буферизованный канало для асинхронной обработки запросов на удаление.
var DeleteChan = make(chan DeleteRequest, 100)

// StartDeleteWorkers запускает указанное количество горутин для обработки запросов на удаление.
// Каждая горутина прослушивает канал DeleteChan и помечает URL как удалённые в БД.
//
// Параметры:
//   - db: подключение к БД для выполнения операций удаления
//   - workerCount: число запускаемых горутин
//
// Функция:
//   - Запускает указанное количество горутин
//   - Каждая горутина слушает DeleteChan на предмет поступающих запросов на удаление
//   - Горутины выполняют SQL-команды обновления для пометки URL как удалённые
//   - Генерирует panic, если возникают ошибки в операциях с базой данных
//
// Горутины:
//   - Обрабатывают объекты DeleteRequest из канала
//   - Обновляют БД, помечая URL как удалённые для указанного пользователя
//
// Пример:
//
//	db := getDatabaseConnection()
//	StartDeleteWorkers(db, 5) // Запустить 5 горутин
//
//	// Отправить запрос на удаление
//	DeleteChan <- DeleteRequest{
//	    UserID:    "user123",
//	    ShortURLs: []string{"abc123", "def456"},
//	}

func StartDeleteWorkers(db *sql.DB, workerCount int) {
	for i := 0; i < workerCount; i++ {
		go func(id int) {
			for req := range DeleteChan {
				_, err := db.Exec(
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

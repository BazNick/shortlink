package entities

import "database/sql"

type DeleteRequest struct {
	UserID    string
	ShortURLs []string
}

var DeleteChan = make(chan DeleteRequest, 100)

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

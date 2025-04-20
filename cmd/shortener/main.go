package main

import (
	"github.com/BazNick/shortlink/cmd/compress"
	"github.com/BazNick/shortlink/cmd/config"
	"github.com/BazNick/shortlink/cmd/logger"
	"github.com/BazNick/shortlink/internal/app/entities"
	"github.com/BazNick/shortlink/internal/app/handlers"
	"github.com/BazNick/shortlink/internal/app/storage"
	"github.com/gin-gonic/gin"
)

func main() {
	var (
		conf    = config.GetCLParams()
		router  = gin.Default()
		storage storage.Storage
	)

	switch {
	case conf.DB != "":
		db := entities.NewDB(conf.DB)
		storage = db

		defer db.Database.Close()
	case conf.FilePath != "":
		file := entities.NewFileStore(conf.FilePath)
		storage = file

		defer file.FileStorage.Close()
	default:
		hashDict := entities.NewHashDict()
		storage = hashDict
	}

	urlHandler := handlers.NewURLHandler(
		storage,
		conf.FilePath,
		conf.DB,
	)

	router.Use(logger.WithLogging(), compress.GzipHandle())

	router.GET("/:id", urlHandler.GetLink)
	router.POST("/", urlHandler.AddLink)
	router.POST("/api/shorten", urlHandler.PostJSONLink)
	router.GET("/ping", urlHandler.DBPingConn)
	router.POST("/api/shorten/batch", urlHandler.BatchLinks)

	router.Run(conf.Address)
}

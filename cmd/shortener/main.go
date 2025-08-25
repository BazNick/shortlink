package main

import (
	"log"
	"runtime"

	"github.com/BazNick/shortlink/cmd/config"
	"github.com/BazNick/shortlink/cmd/middleware/auth"
	"github.com/BazNick/shortlink/cmd/middleware/compress"
	"github.com/BazNick/shortlink/cmd/middleware/logger"
	"github.com/BazNick/shortlink/internal/app/entities"
	"github.com/BazNick/shortlink/internal/app/handlers"
	"github.com/BazNick/shortlink/internal/app/storage"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

func main() {
	var (
		conf    config.Config
		router  = gin.Default()
		storage storage.Storage
		err     error
	)

	conf, err = config.GetCLParams()
	if err != nil {
		log.Fatal(err)
	}

	pprof.Register(router)

	switch {
	case conf.DB != "":
		db := entities.NewDB(conf.DB)
		storage = db

		entities.StartDeleteWorkers(db.Database, runtime.NumCPU())

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

	router.Use(
		logger.WithLogging(),
		logger.WithLogging(),
		compress.GzipHandle(),
		auth.Auth(conf.SecretKey),
	)

	router.GET("/:id", urlHandler.GetLink)
	router.POST("/", urlHandler.AddLink)
	router.POST("/api/shorten", urlHandler.PostJSONLink)
	router.GET("/ping", urlHandler.DBPingConn)
	router.POST("/api/shorten/batch", urlHandler.BatchLinks)
	router.GET("/api/user/urls", urlHandler.GetUserLinks)
	router.DELETE("/api/user/urls", urlHandler.DeleteUserLinks)

	router.Run(conf.Address)
}

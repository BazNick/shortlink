package main

import (
	"github.com/BazNick/shortlink/cmd/compress"
	"github.com/BazNick/shortlink/cmd/config"
	"github.com/BazNick/shortlink/cmd/logger"
	"github.com/BazNick/shortlink/internal/app/entities"
	"github.com/BazNick/shortlink/internal/app/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
	conf := config.GetCLParams()

	router := gin.Default()
	router.Use(logger.WithLogging(), compress.GzipHandle())

	hashDict := entities.NewHashDict()
	urlHandler := handlers.NewURLHandler(
		hashDict, 
		conf.FilePath,
		conf.DB,
	)

	router.GET("/:id", urlHandler.GetLink)
	router.POST("/", urlHandler.AddLink)
	router.POST("/api/shorten", urlHandler.PostJSONLink)
	router.GET("/ping", urlHandler.DBPingConn)

	router.Run(conf.Address)
}

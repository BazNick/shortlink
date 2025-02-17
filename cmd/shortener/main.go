package main

import (
	"github.com/BazNick/shortlink/cmd/config"
	"github.com/BazNick/shortlink/internal/app/handlers"
	"github.com/BazNick/shortlink/internal/app/entities"
	"github.com/gin-gonic/gin"
)

func main() {
	conf := config.GetCLParams()

	router := gin.Default()

	hashDict := entities.NewHashDict()
	urlHandler := handlers.NewURLHandler(hashDict)

	router.GET("/:id", urlHandler.GetLink)
	router.POST("/", urlHandler.AddLink)

	router.Run(conf.Address)
}

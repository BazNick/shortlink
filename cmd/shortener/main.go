package main

import (
	"github.com/BazNick/shortlink/cmd/config"
	"github.com/BazNick/shortlink/internal/app/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
	conf := config.GetCLParams()

	router := gin.Default()

	router.GET("/:id", handlers.GetLink)
	router.POST("/", handlers.AddLink)

	router.Run(conf.Address)
}

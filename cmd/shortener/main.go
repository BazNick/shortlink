package main

import (
	"github.com/BazNick/shortlink/internal/app/handlers"
	"github.com/gin-gonic/gin"
)


func main() {
	router := gin.Default()

    router.GET("/:id", handlers.GetLink)
    router.POST("/", handlers.AddLink)

    router.Run(":8080")
}

package auth

import "github.com/gin-gonic/gin"

func AuthHandle() gin.HandlerFunc {
    return func(c *gin.Context) {
        
        c.Next()
    }
}
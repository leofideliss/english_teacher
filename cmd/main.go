package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/leofideliss/english_teacher/internal"
)

func main() {
    // LOAD .env
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    
    router := gin.Default()
    router.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "pong",
        })
    })

    router.POST("/question", internal.ExecuteQuestion)
    
    router.Run()
}

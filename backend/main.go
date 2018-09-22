package main

import (
	"./api"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.Use(cors.Default())
	r.Static("/login", "./app")
	r.GET("/challenge", api.ChallengeGet)
	r.POST("/challenge", api.ChallengePost)
	r.Run() // listen and serve on 0.0.0.0:8080
}

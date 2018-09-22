package main

import (
	"./api"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://127.0.0.1:4200"}
	r.Use(cors.New(config))
	r.GET("/challenge", api.ChallengeGet)
	r.POST("/challenge", api.ChallengePost)
	r.Run() // listen and serve on 0.0.0.0:8080
}

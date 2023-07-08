package main

import (
	"fmt"
	"natwin/config"
	"natwin/server"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")

	server.AddRouter(router)
	router.Static("/static", "./static")
	router.StaticFile("/favicon.ico", "./static/favicon.ico")
	
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, fmt.Sprint(time.Now().Unix()))
	})

	router.Run(config.Listen)
}

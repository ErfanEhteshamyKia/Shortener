package main

import "github.com/gin-gonic/gin"

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.POST("/shorten", shorten)
	r.GET("/redirect/:shorthand", redirect)
	r.GET("/analytics/:shorthand", analytics)
	return r
}

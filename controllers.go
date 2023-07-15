package main

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

func shorten(c *gin.Context) {
	var requestBody URL
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "json decoding error"})
		return
	}

	if requestBody.Long == "" || requestBody.Short == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "please provide both long and short url"})
		return
	}
	_, err := collection.InsertOne(context.TODO(), requestBody)
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusCreated, gin.H{"status": true})
}

func redirect(c *gin.Context) {}

func analytics(c *gin.Context) {}

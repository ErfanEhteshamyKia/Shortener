package main

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
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

func redirect(c *gin.Context) {
	shorthand := c.Param("shorthand")
	result := collection.FindOne(context.TODO(), bson.D{{"short", shorthand}})

	var url URL
	result.Decode(&url)

	c.Redirect(http.StatusMovedPermanently, url.Long)
}

func analytics(c *gin.Context) {}

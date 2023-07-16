package main

import (
	"context"
	"net/http"
	"strings"

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

	// Store password hash in db
	if requestBody.Password != "" {
		requestBody.Password = hash(requestBody.Password)
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

	if url.Password == "" {
		c.Redirect(http.StatusMovedPermanently, url.Long)
		return
	}

	auth_header := c.GetHeader("Authorization")
	password := strings.Split(auth_header, " ")[1] // Bearer token
	if compare_password(password, url.Password) {
		c.Redirect(http.StatusMovedPermanently, url.Long)
		return
	}
	c.JSON(http.StatusUnauthorized, gin.H{"status": false, "message": "please provide password in the form of bearer token viewing this url requires authentication"})
}

func analytics(c *gin.Context) {}

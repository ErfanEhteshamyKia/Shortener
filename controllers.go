package main

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func shorten(c *gin.Context) {
	var requestBody URL
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"success": false, "message": "json decoding error"})
	}

	if requestBody.Long == "" || requestBody.Short == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"success": false, "message": "please provide both long and short url"})
	}

	if requestBody.ExpiredAt.IsZero() {
		requestBody.ExpiredAt = time.Now().Add(24 * time.Hour)
	}

	// Store password hash in db
	if requestBody.Password != "" {
		requestBody.Password = hash(requestBody.Password)
	}

	result := collection.FindOne(context.TODO(), bson.D{{"short", requestBody.Short}})
	if result.Err() != mongo.ErrNoDocuments {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"success": false, "message": "shorthand already exists"})
	}

	_, err := collection.InsertOne(context.TODO(), requestBody)
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusCreated, gin.H{"success": true})
}

func redirect(c *gin.Context) {
	shorthand := c.Param("shorthand")
	result := collection.FindOne(context.TODO(), bson.D{{"short", shorthand}})
	if result.Err() == mongo.ErrNoDocuments {
		c.AbortWithStatus(http.StatusNotFound)
	}

	var url URL
	result.Decode(&url)

	now := time.Now()
	if now.After(url.ExpiredAt) {
		c.AbortWithStatus(http.StatusGone)
	}

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
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "message": "please provide password in the form of bearer token viewing this url requires authentication"})
}

func analytics(c *gin.Context) {}

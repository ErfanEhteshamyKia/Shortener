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

	auth_header := c.GetHeader("Authorization")
	password := ""
	if len(strings.Split(auth_header, " ")) == 2 {
		password = strings.Split(auth_header, " ")[1] // Bearer token
	}
	if url.Password == "" || compare_password(password, url.Password) {
		click := Click{Short: shorthand, Time: time.Now()}
		collection.InsertOne(context.TODO(), click)
		c.Redirect(http.StatusMovedPermanently, url.Long)
		return
	}
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "message": "please provide password in the form of bearer token viewing this url requires authentication"})
}

func analytics(c *gin.Context) {
	shorthand := c.Param("shorthand")
	var requestBody AnalyticsRequest
	c.ShouldBindJSON(&requestBody)

	if requestBody.Finish.IsZero() {
		requestBody.Finish = time.Now()
	}

	filter := bson.M{
		"short": shorthand,
		"time": bson.M{
			"$gt": requestBody.Start,
			"$lt": requestBody.Finish,
		},
	}
	clicks, err := collection.CountDocuments(context.TODO(), filter)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"success": false, "message": "database access error"})
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "clicks": clicks})
}

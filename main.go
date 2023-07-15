package main

import (
	"go.mongodb.org/mongo-driver/mongo"
)

var collection *mongo.Collection

func main() {
	cancel, disconnect := setupDB(false)
	defer cancel()
	defer disconnect()

	r := setupRouter()
	r.Run()
}

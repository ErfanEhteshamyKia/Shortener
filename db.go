package main

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func setupDB(is_test bool) (context.CancelFunc, func()) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}
	disconnect := func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}

	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		panic(err)
	}

	if is_test {
		collection = client.Database("test_shortener").Collection("test_urls")
	} else {
		collection = client.Database("shortener").Collection("urls")
	}
	return cancel, disconnect
}

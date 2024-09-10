package models

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Connect() (*mongo.Client, error) {

	mongo_url := os.Getenv("MONGODB_URL")

	// Set client options
	clientOptions := options.Client().ApplyURI(mongo_url)
	// Increase the timeout
	clientOptions = clientOptions.SetConnectTimeout(60 * time.Second)
	clientOptions = clientOptions.SetServerSelectionTimeout(60 * time.Second)

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), clientOptions)
    if err != nil {
        return nil, err
    }

	// Check the connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		 log.Fatal(err)
	}
	return client, nil
}
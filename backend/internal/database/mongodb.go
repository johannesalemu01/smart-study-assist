package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client

func ConnectDB() *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	clientOptions := options.Client().ApplyURI(mongoURI).
		SetMaxPoolSize(50).
		SetMinPoolSize(5).
		SetMaxConnIdleTime(5 * time.Minute).
		SetServerSelectionTimeout(5 * time.Second).
		SetConnectTimeout(10 * time.Second)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %v", err)
	}

	// Ping the database
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Error pinging MongoDB: %v", err)
	}

	fmt.Println("Connected to MongoDB!")
	Client = client
	return client
}

func GetCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "smart_study_assist"
	}
	return client.Database(dbName).Collection(collectionName)
}

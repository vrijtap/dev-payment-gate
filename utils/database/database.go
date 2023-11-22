package database

import (
	"context"
	"fmt"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	c		*mongo.Client
    db		*mongo.Database
    dbLock	sync.Mutex
)

// Connect initializes the MongoDB client and establishes a connection
func Connect(uri, database string) error {
	// Check if the client is already connected
	if c != nil {
		// Disconnect the connection
        if err := Disconnect(); err != nil {
			return err
		}
    }

    // Use the SetServerAPIOptions() method to set the Stable API version to 1
    apiOptions := options.ServerAPI(options.ServerAPIVersion1)
    clientOptions := options.Client().ApplyURI(uri).SetServerAPIOptions(apiOptions)

    // Connect to MongoDB
    client, err := mongo.Connect(context.Background(), clientOptions)
    if err != nil {
        return err
    }

    // Ping MongoDB to verify the connection
    err = client.Ping(context.Background(), nil)
    if err != nil {
        return err
    }

	// Lock the mutex to safely set the client and database
    dbLock.Lock()
    defer dbLock.Unlock()

    // Set the client and database
	c = client
    db = client.Database(database)

    // Log the initialization
    fmt.Println("Connected to MongoDB")

    return nil
}

// Disconnect clears the initialized MongoDB client and closes the connection
func Disconnect() error {
    // Check if the client is already closed
    if c == nil {
        return nil
    }

	// Lock the mutex to safely clear the client and database
    dbLock.Lock()
    defer dbLock.Unlock()

    // Disconnect from MongoDB
    err := c.Disconnect(context.Background())
    if err != nil {
        return err
    }

	// Clear the client and database
    c = nil
	db = nil

    return nil
}

// Get returns a collection in the database
func GetCollection(collection string) *mongo.Collection {
	// Lock the mutex to safely get the database
    dbLock.Lock()
    defer dbLock.Unlock()

	// Return the database
	return db.Collection(collection)
}

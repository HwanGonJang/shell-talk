package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// NewDB creates a new MongoDB client and connects to the database.
func NewDB(ctx context.Context, connectionString string) (*mongo.Database, error) {
	clientOptions := options.Client().ApplyURI(connectionString)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongo: %w", err)
	}

	// Ping the primary to verify the connection.
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping mongo: %w", err)
	}

	// The database name can be part of the connection string or specified here.
	// For simplicity, we'll assume a database name, e.g., "shelltalk".
	db := client.Database("shelltalk")

	return db, nil
}

package main

import (
	"context"
	"database/sql"
	"shell-talk-server/internal/config"
	"shell-talk-server/internal/repository/mongo"
	"shell-talk-server/internal/repository/postgres"

	mongodriver "go.mongodb.org/mongo-driver/mongo"
)

func providePostgresDB(cfg *config.Config) (*sql.DB, func(), error) {
	db, err := postgres.NewDB(cfg.PostgresURL)
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() { db.Close() }
	return db, cleanup, nil
}

func provideMongoDB(ctx context.Context, cfg *config.Config) (*mongodriver.Database, func(), error) {
	db, err := mongo.NewDB(ctx, cfg.MongoURL)
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() { db.Client().Disconnect(ctx) }
	return db, cleanup, nil
}

func provideContext() (context.Context, func()) {
	ctx, cancel := context.WithCancel(context.Background())
	return ctx, func() { cancel() }
}

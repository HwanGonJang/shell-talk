//go:build wireinject
// +build wireinject

package main

import (
	"shell-talk-server/internal/config"
	"shell-talk-server/internal/hub"
	"shell-talk-server/internal/repository/mongo"
	"shell-talk-server/internal/repository/postgres"
	"shell-talk-server/internal/service"

	"github.com/google/wire"
)

// App is the main application container.
type App struct {
	Hub *hub.Hub
}

// InitializeApp creates a new application.
func InitializeApp() (*App, func(), error) {
	wire.Build(
		config.Load,
		// Database & Context Providers
		wire.NewSet(
			provideContext,
			providePostgresDB,
			provideMongoDB,
		),
		// Repository Providers
		wire.NewSet(
			postgres.NewUserRepository,
			wire.Bind(new(service.IUserRepository), new(*postgres.UserRepository)),

			postgres.NewRoomRepository,
			wire.Bind(new(service.IRoomRepository), new(*postgres.RoomRepository)),

			mongo.NewMessageRepository,
			wire.Bind(new(service.IMessageRepository), new(*mongo.MessageRepository)),
		),
		// Service Providers
		wire.NewSet(
			service.NewUserService,
			wire.Bind(new(service.IUserService), new(*service.UserService)),

			service.NewRoomService,
			wire.Bind(new(service.IRoomService), new(*service.RoomService)),
		),
		// Hub Provider
		hub.NewHub,
		// App Provider
		wire.NewSet(
			wire.Struct(new(App), "*"),
		),
	)
	return nil, nil, nil
}

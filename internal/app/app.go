package app

import (
	grpcapp "catalogue-service/internal/app/grpc"
	"catalogue-service/internal/data"
	"catalogue-service/internal/services/catalogue"
	"log/slog"
	"time"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(
	log *slog.Logger,
	grpcPort int,
	dsn string,
	tokenTTL time.Duration,
) *App {
	// TODO: database setup
	itemRepo, err := data.New(dsn)
	if err != nil {
		panic(err)
	}

	// TODO: catalogue service setup in services/catalogue
	catalogueService := catalogue.New(log, itemRepo, tokenTTL)

	// TODO: grpc app setup
	grpcApp := grpcapp.New(log, catalogueService, grpcPort)

	return &App{GRPCServer: grpcApp}
}

package main

import (
	"context"
	"os"
	"time"

	"github.com/jkaninda/goma-admin/config"
	"github.com/jkaninda/goma-admin/routes"
	"github.com/jkaninda/goma-admin/store"
	"github.com/jkaninda/logger"
	"github.com/jkaninda/okapi"
	"github.com/jkaninda/okapi/okapicli"
)

func main() {
	app := okapi.New()
	cli := okapicli.New(app, "Goma").
		String("config", "c", "config.yaml", "Path to configuration file").
		Int("port", "p", 8080, "HTTP server port")
	conf, err := config.New(app, cli)
	if err != nil {
		logger.Fatal("Failed to initialize config", "error", err)

	}
	if err := store.AutoMigrate(conf.Database.DB); err != nil {
		logger.Fatal("Failed to run migrations", "error", err)
	}

	// Create the route instance
	route := routes.NewRoute(context.Background(), app, conf)
	// Register routes
	route.RegisterRoutes()
	// Start the server
	if err := cli.RunServer(&okapicli.RunOptions{
		ShutdownTimeout: 30 * time.Second,                               // Optional: customize shutdown timeout
		Signals:         []os.Signal{okapicli.SIGINT, okapicli.SIGTERM}, // Optional: customize shutdown signals
		OnStart: func() {
			logger.Info("Preparing resources before startup")

		},
		OnStarted: func() {
			logger.Info("Server started successfully")
		},
		OnShutdown: func() {
			logger.Info("Cleaning up before shutdown")
		},
	}); err != nil {
		panic(err)
	}
}

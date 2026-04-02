package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"

	"github.com/alperkirkus/fintech-backend/internal/config"
	"github.com/alperkirkus/fintech-backend/internal/database"
	"github.com/alperkirkus/fintech-backend/internal/logger"
)

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("no .env file found, reading from environment")
	}

	configuration, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	logger.Init(configuration.Log.Level, configuration.Log.Format)

	log.Info().
		Str("env", configuration.App.Env).
		Str("port", configuration.App.Port).
		Msg("starting server")

	databaseConnection, err := database.Connect(configuration.DB)
	if err != nil {
		log.Fatal().Err(err).Msg("database setup failed")
	}
	defer databaseConnection.Close()

	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/health", healthHandler)

	httpServer := &http.Server{
		Addr:         ":" + configuration.App.Port,
		Handler:      serveMux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErrorChannel := make(chan error, 1)
	go func() {
		log.Info().Str("addr", httpServer.Addr).Msg("http server listening")
		serverErrorChannel <- httpServer.ListenAndServe()
	}()

	shutdownSignalChannel := make(chan os.Signal, 1)
	signal.Notify(shutdownSignalChannel, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrorChannel:
		log.Error().Err(err).Msg("server error")
	case receivedSignal := <-shutdownSignalChannel:
		log.Info().Str("signal", receivedSignal.String()).Msg("shutdown signal received")
	}

	shutdownContext, cancelShutdown := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelShutdown()

	if err := httpServer.Shutdown(shutdownContext); err != nil {
		log.Error().Err(err).Msg("forced shutdown")
	}

	log.Info().Msg("server stopped gracefully")
}

func healthHandler(responseWriter http.ResponseWriter, request *http.Request) {
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(http.StatusOK)
	fmt.Fprintln(responseWriter, `{"status":"ok"}`)
}

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

	"github.com/alperkirkus/fintech-backend/internal/api"
	"github.com/alperkirkus/fintech-backend/internal/api/handler"
	"github.com/alperkirkus/fintech-backend/internal/config"
	"github.com/alperkirkus/fintech-backend/internal/database"
	"github.com/alperkirkus/fintech-backend/internal/logger"
	"github.com/alperkirkus/fintech-backend/internal/middleware"
	"github.com/alperkirkus/fintech-backend/internal/model"
	"github.com/alperkirkus/fintech-backend/internal/service"
	pgstore "github.com/alperkirkus/fintech-backend/internal/store/postgres"
)

func main() {
	serverCtx, serverCancel := context.WithCancel(context.Background())
	defer serverCancel()

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

	db, err := database.Connect(configuration.DB)
	if err != nil {
		log.Fatal().Err(err).Msg("database setup failed")
	}
	defer db.Close()

	userStore := pgstore.NewUserStore(db.DB)
	transactionStore := pgstore.NewTransactionStore(db.DB)
	balanceStore := pgstore.NewBalanceStore(db.DB)

	authService := service.NewAuthService(userStore, configuration.JWT.Secret, configuration.JWT.TTL)
	userService := service.NewUserService(userStore)
	transactionService := service.NewTransactionService(db.DB, transactionStore, 10)
	balanceService := service.NewBalanceService(balanceStore, transactionStore)

	authHandler := handler.NewAuthHandler(authService, userService)
	userHandler := handler.NewUserHandler(userService)
	transactionHandler := handler.NewTransactionHandler(transactionService, balanceService)
	balanceHandler := handler.NewBalanceHandler(balanceService)

	authMiddleware := middleware.Auth(authService)
	adminMiddleware := middleware.RequireRole(model.RoleAdmin)

	corsConfig := middleware.DefaultCORSConfig()
	if len(configuration.CORS.AllowedOrigins) > 0 {
		corsConfig.AllowedOrigins = configuration.CORS.AllowedOrigins
	}

	rateLimitConfig := middleware.RateLimitConfig{
		Enabled:           configuration.RateLimit.Enabled,
		RequestsPerSecond: configuration.RateLimit.RequestsPerSecond,
		Burst:             configuration.RateLimit.Burst,
	}

	router := api.NewRouter()
	router.Use(middleware.Recover())
	router.Use(middleware.Logger())
	router.Use(middleware.Performance())
	router.Use(middleware.Security())
	router.Use(middleware.CORS(corsConfig))
	router.Use(middleware.Validate())
	router.Use(middleware.RateLimit(serverCtx, rateLimitConfig))

	router.HandleFunc("GET /health", healthHandler)
	router.HandleFunc("GET /metrics", middleware.MetricsHandler())

	router.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	router.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	router.HandleFunc("POST /api/v1/auth/refresh", authHandler.Refresh, authMiddleware)

	router.HandleFunc("GET /api/v1/users", userHandler.List, authMiddleware, adminMiddleware)
	router.HandleFunc("GET /api/v1/users/{id}", userHandler.GetByID, authMiddleware)
	router.HandleFunc("PUT /api/v1/users/{id}", userHandler.Update, authMiddleware)
	router.HandleFunc("DELETE /api/v1/users/{id}", userHandler.Delete, authMiddleware)

	router.HandleFunc("POST /api/v1/transactions/credit", transactionHandler.Credit, authMiddleware)
	router.HandleFunc("POST /api/v1/transactions/debit", transactionHandler.Debit, authMiddleware)
	router.HandleFunc("POST /api/v1/transactions/transfer", transactionHandler.Transfer, authMiddleware)
	router.HandleFunc("GET /api/v1/transactions/history", transactionHandler.History, authMiddleware)
	router.HandleFunc("GET /api/v1/transactions/{id}", transactionHandler.GetByID, authMiddleware)

	router.HandleFunc("GET /api/v1/balances/current", balanceHandler.Current, authMiddleware)
	router.HandleFunc("GET /api/v1/balances/historical", balanceHandler.Historical, authMiddleware)
	router.HandleFunc("GET /api/v1/balances/at-time", balanceHandler.AtTime, authMiddleware)

	httpServer := &http.Server{
		Addr:         ":" + configuration.App.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErrorChannel := make(chan error, 1)
	go func() {
		log.Info().Str("addr", httpServer.Addr).Msg("http server listening")
		serverErrorChannel <- httpServer.ListenAndServe()
	}()

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrorChannel:
		log.Error().Err(err).Msg("server error")
	case sig := <-signalChannel:
		log.Info().Str("signal", sig.String()).Msg("shutdown signal received")
	}

	serverCancel()

	shutdownContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownContext); err != nil {
		log.Error().Err(err).Msg("forced shutdown")
	}

	log.Info().Msg("server stopped gracefully")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, `{"status":"ok"}`)
}

package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"shouketsu/internal/entities"
	"shouketsu/internal/handlers/ws"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	router := gin.Default()

	err := godotenv.Load()
	if err != nil {
		slog.Error("No .env file found")
	}

	// logger
	var logger *slog.Logger

	if os.Getenv("APP_ENV") == "prod" {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	}

	slog.SetDefault(logger)

	slog.Info("Logger service has started", slog.String("env", os.Getenv("APP_ENV")))

	PORT := os.Getenv("PORT")

	room := entities.NewRoom("room_test")
	go ws.RunRoom(room)

	wsHandler := ws.NewWShandler(room)

	router.GET("/ws", wsHandler.ServeWS)

	srv := &http.Server{
		Addr:    ":" + PORT,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("listen and serve failed", slog.Any("err", err))
			os.Exit(1)
		}
	}()
	slog.Info("server started succesfully", slog.String("port", PORT))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	slog.Info("shutting down server...")

	shutdownTime := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTime)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", slog.Any("err", err))
		os.Exit(1)
	}
	slog.Info("server exited cleanly")
}

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zepheo/guild_manager/internal/ingestor/csv"
	"github.com/zepheo/guild_manager/internal/ingestor/turtlogs"
	"github.com/zepheo/guild_manager/internal/service"
	"github.com/zepheo/guild_manager/internal/storage/postgres"
)

func main() {
	// 1. Connect to DB
	dbURL := os.Getenv("DB_URL")
	db, err := postgres.Connect(dbURL)
	if err != nil {
		log.Fatal("Could not connect to DB:", err)
	}

	// 2. Run Migrations
	if err := postgres.RunMigrations(db); err != nil {
		log.Fatal("Migration failed:", err)
	}

	log.Println("Database is ready. Starting API on :8080...")

	repo := &postgres.PostgresRaidRepo{Db: db}
	parser := csv.NewParser()
	client := &turtlogs.TurtlogsClient{BaseURL: "https://turtlogs.com"}

	raidService := service.NewRaidService(repo, parser, client)

	handler := NewRaidHandler(raidService)

	// 3. Setup API Router
	router := gin.Default()
	handler.RegisterRoutes(router)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	log.Println("Server started on :8080")

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}

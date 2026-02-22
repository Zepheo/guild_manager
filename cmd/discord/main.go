package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	discordBot "github.com/zepheo/guild_manager/internal/discord"
	"github.com/zepheo/guild_manager/internal/ingestor/csv"
	"github.com/zepheo/guild_manager/internal/ingestor/turtlogs"
	"github.com/zepheo/guild_manager/internal/service"
	"github.com/zepheo/guild_manager/internal/storage/postgres"
)

func main() {
	// 1. Database Connection
	db, err := postgres.Connect(os.Getenv("DB_URL"))
	if err != nil {
		log.Fatalf("Bot failed to connect to DB: %v", err)
	}

	if err := postgres.RunMigrations(db); err != nil {
		log.Fatal("Migration failed:", err)
	}

	repo := &postgres.PostgresRaidRepo{Db: db}

	// 2. Initialize Bot
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN environment variable is required")
	}

	parser := csv.NewParser()

	baseUrl := os.Getenv("LOGS_URL")
	if baseUrl == "" {
		log.Fatalf("LOGS_URL environment variable is required")
	}
	client := turtlogs.NewTurtlogsClient(baseUrl)

	service := service.NewRaidService(repo, parser, client)
	b, err := discordBot.NewBot(token, service, repo)
	if err != nil {
		log.Fatalf("Failed to initialize bot: %v", err)
	}

	// 3. Register Slash Commands & Handlers
	b.RegisterHandlers()
	if err := b.Session.Open(); err != nil {
		log.Fatalf("Error opening Discord session: %v", err)
	}

	_, err = b.RegisterCommands()
	if err != nil {
		log.Fatalf("Error registering commands: %v", err)
	}
	defer b.Session.Close()

	log.Println("Discord Bot is now running...")

	// 4. Wait for termination
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down bot gracefully...")
}

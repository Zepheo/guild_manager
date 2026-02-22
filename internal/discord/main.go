package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/zepheo/guild_manager/internal/service"
	"github.com/zepheo/guild_manager/internal/storage/postgres"
)

type DiscordBot struct {
	Session     *discordgo.Session
	RaidService *service.RaidService
	RaidRepo    *postgres.PostgresRaidRepo
}

func NewBot(token string, svc *service.RaidService, repo *postgres.PostgresRaidRepo) (*DiscordBot, error) {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	return &DiscordBot{
		Session:     dg,
		RaidService: svc,
		RaidRepo:    repo,
	}, nil
}

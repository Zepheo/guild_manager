package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/zepheo/guild_manager/internal/service"
)

type Bot struct {
	Session     *discordgo.Session
	RaidService *service.RaidService
}

func NewBot(token string, svc *service.RaidService) (*Bot, error) {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	return &Bot{
		Session:     dg,
		RaidService: svc,
	}, nil
}

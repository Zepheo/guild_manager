package discord

import (
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/zepheo/guild_manager/internal/domain"
)

var Commands = []*discordgo.ApplicationCommand{
	{
		Name:        "score",
		Description: "Check current SR+ bonus",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionAttachment,
				Name:        "file",
				Description: "Reserve list to update SR for",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "raid",
				Description: "Raid to update SR for",
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{
						Name:  "MC",
						Value: domain.MC,
					},
					{
						Name:  "ES",
						Value: domain.ES,
					},
					{
						Name:  "BWL",
						Value: domain.BWL,
					},
				},
			},
		},
	},
	{
		Name:        "process",
		Description: "Admin: Process a Turtlogs link and CSV file",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "url",
				Description: "Turtlogs URL",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionAttachment,
				Name:        "file",
				Description: "The CSV file containing reserves",
				Required:    true,
			},
		},
	},
}

func (b *DiscordBot) RegisterCommands() ([]*discordgo.ApplicationCommand, error) {
	guildID := os.Getenv("DISCORD_GUILD_ID")

	if guildID == "" {
		log.Fatal("DISCORD_GUILD_ID environment variable is not set")
	}

	registeredCommands := make([]*discordgo.ApplicationCommand, len(Commands))
	for i, c := range Commands {
		cmd, err := b.Session.ApplicationCommandCreate(b.Session.State.User.ID, guildID, c)
		if err != nil {
			return nil, err
		}
		registeredCommands[i] = cmd
	}
	return registeredCommands, nil
}

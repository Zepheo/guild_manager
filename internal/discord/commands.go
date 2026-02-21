package discord

import "github.com/bwmarrin/discordgo"

var Commands = []*discordgo.ApplicationCommand{
	{
		Name:        "score",
		Description: "Check current SR+ bonus",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "character",
				Description: "Character name",
				Required:    true,
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

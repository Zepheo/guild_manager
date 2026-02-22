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

func (b *DiscordBot) RegisterCommands() ([]*discordgo.ApplicationCommand, error) {
	registeredCommands := make([]*discordgo.ApplicationCommand, len(Commands))
	for i, c := range Commands {
		cmd, err := b.Session.ApplicationCommandCreate(b.Session.State.User.ID, "960543410530418708", c)
		if err != nil {
			return nil, err
		}
		registeredCommands[i] = cmd
	}
	return registeredCommands, nil
}

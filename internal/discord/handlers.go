package discord

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bwmarrin/discordgo"
)

func (b *DiscordBot) RegisterHandlers() {
	b.Session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionApplicationCommand {
			return
		}

		switch i.ApplicationCommandData().Name {
		case "score":
			b.handleScore(s, i)
		case "process":
			b.handleProcess(s, i)
		}
	})
}

func (b *DiscordBot) handleProcess(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	options := i.ApplicationCommandData().Options
	log := options[0].StringValue()

	attachmentID := options[1].Value.(string)
	attachment := i.ApplicationCommandData().Resolved.Attachments[attachmentID]

	resp, err := http.Get(attachment.URL)
	if err != nil {
		b.respondError(s, i, "Failed to download CSV from Discord")
		return
	}
	defer resp.Body.Close()

	err = b.RaidService.ProcessRaid(context.Background(), resp.Body, log)
	if err != nil {
		b.respondError(s, i, fmt.Sprintf("Processing failed: %v", err))
		return
	}

	content := fmt.Sprintf("✅ Successfully processed raid **%s**. Scores updated!", log)
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
}

func (b *DiscordBot) handleScore(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
	})

	options := i.ApplicationCommandData().Options
	characterName := options[0].StringValue()

	b.RaidRepo.GetPlayerBonus(context.Background(), characterName)
}

func (b *DiscordBot) respondError(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) {
	errContent := "❌ " + msg
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &errContent,
	})
}

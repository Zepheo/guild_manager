package discord

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/zepheo/guild_manager/internal/domain"
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
		Data: &discordgo.InteractionResponseData{
			Content: "⏳ Calculating SR+ bonuses for the raid...",
		},
	})

	options := i.ApplicationCommandData().Options

	attachementID := options[0].Value.(string)
	attachement := i.ApplicationCommandData().Resolved.Attachments[attachementID]
	raid := options[1].Value.(domain.Raid)

	resp, err := http.Get(attachement.URL)
	if err != nil {
		fmt.Println("error here")
		b.respondError(s, i, "Failed to download CSV from Discord")
		return
	}
	defer resp.Body.Close()

	results, err := b.RaidService.CalculateSRPlus(context.Background(), resp.Body, raid)
	if err != nil {
		b.respondError(s, i, fmt.Sprintf("Calculation failed: %v", err))
		return
	}

	players := make([]string, 0, len(results))
	for player := range results {
		players = append(players, player)
	}
	sort.Strings(players)

	var sb strings.Builder
	sb.WriteString("📊 **SR+ Bonuses**:\n")

	hasContent := false
	for _, player := range players {
		items := results[player]
		if len(items) > 0 {
			hasContent = true
			sb.WriteString(fmt.Sprintf("👤 **%s**\n", player))
			for _, item := range items {
				sb.WriteString(fmt.Sprintf("- %s: **%d**\n", item.ItemName, item.SR))
			}
			sb.WriteString("\n")
		}
	}

	content := sb.String()
	if !hasContent {
		content = "ℹ️ No bonuses found for the players in this CSV."
	}

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
}

func (b *DiscordBot) respondError(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) {
	errContent := "❌ " + msg
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &errContent,
	})
}

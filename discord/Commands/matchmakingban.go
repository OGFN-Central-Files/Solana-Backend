package commands

import (
	"fmt"
	"solana/modules/database"
	"solana/modules/database/entities/accounts"
	"time"

	"github.com/bwmarrin/discordgo"
)

const GuildIDD = "1334285164028887071"

// idc if this is bad you can fix it idgaf atp
var predefinedReasons = []string{
	"Harassment",
	"TradeScam",
	"AFK",
	"Exploiting",
	"Offensive",
}

type MatchmakingBanCommand struct {
	Name        string
	Description string
	Options     []*discordgo.ApplicationCommandOption
}

func NewMatchmakingBanCommand() *MatchmakingBanCommand {

	reasonChoices := make([]*discordgo.ApplicationCommandOptionChoice, len(predefinedReasons))
	for i, reason := range predefinedReasons {
		reasonChoices[i] = &discordgo.ApplicationCommandOptionChoice{
			Name:  reason,
			Value: reason,
		}
	}

	return &MatchmakingBanCommand{
		Name:        "matchmakingban",
		Description: "Bans a user from matchmaking",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "The user to ban",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "ban_duration",
				Description: "Time the user will be banned for (days)",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "reason",
				Description: "The reason for the ban",
				Required:    true,
				Choices:     reasonChoices,
			},
		},
	}
}

func (c *MatchmakingBanCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options

	var userOption *discordgo.User
	var banDuration int64
	var reasonOption string

	for _, opt := range options {
		switch opt.Name {
		case "user":
			userOption = opt.UserValue(s)
		case "ban_duration":
			banDuration = opt.IntValue()
		case "reason":
			reasonOption = opt.StringValue()
		}
	}

	if userOption == nil || banDuration <= 0 || reasonOption == "" {
		respondWithMatchmakingError(s, i, "Invalid parameters provided.")
		return
	}

	validReason := false
	for _, reason := range predefinedReasons {
		if reasonOption == reason {
			validReason = true
			break
		}
	}

	if !validReason {
		respondWithMatchmakingError(s, i, "Invalid reason provided.")
		return
	}

	db := database.Get()
	var userAccount accounts.Users
	result := db.Where(`"discordId" = ?`, userOption.ID).First(&userAccount)
	if result.Error != nil {
		respondWithMatchmakingError(s, i, "This user doesn't have an account!")
		return
	}

	if userAccount.MatchmakingBannedUntil != "" {
		respondWithMatchmakingError(s, i, fmt.Sprintf("User `%s` is already banned from matchmaking!", userAccount.DisplayName))
		return
	}

	banUntil := time.Now().Add(time.Duration(banDuration) * 24 * time.Hour).Format(time.RFC3339)
	banSince := time.Now().Format(time.RFC3339)

	userAccount.MatchmakingBannedUntil = banUntil
	userAccount.MatchmakingBannedSince = banSince
	userAccount.MatchmakingBannedReason = reasonOption

	result = db.Model(&userAccount).Updates(map[string]interface{}{
		"matchmakingBannedUntil":  banUntil,
		"matchmakingBannedSince":  banSince,
		"matchmakingBannedReason": reasonOption,
	})
	if result.Error != nil {
		respondWithMatchmakingError(s, i, "An error occurred while processing the ban.")
		return
	}

	dmChannel, err := s.UserChannelCreate(userOption.ID)
	if err == nil {
		_, _ = s.ChannelMessageSend(dmChannel.ID, fmt.Sprintf("You have been banned from matchmaking for %d days. Reason: `%s`", banDuration, reasonOption))
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Banned `%s` from matchmaking for %d days. Reason: `%s`", userAccount.DisplayName, banDuration, reasonOption),
		},
	})
}

func respondWithMatchmakingError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

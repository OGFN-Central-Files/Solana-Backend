package commands

import (
	"fmt"
	"solana/modules/database"
	"solana/modules/database/entities/accounts"

	"github.com/bwmarrin/discordgo"
)

const GuildID = "1334285164028887071"

type BanCommand struct {
	Name        string
	Description string
	Options     []*discordgo.ApplicationCommandOption
}

func NewBanCommand() *BanCommand {
	return &BanCommand{
		Name:        "ban",
		Description: "Bans a user's account",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "The user to ban",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "reason",
				Description: "The reason for the ban",
				Required:    true,
			},
		},
	}
}

func (c *BanCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options

	var userOption *discordgo.User
	var reasonOption string

	for _, opt := range options {
		switch opt.Name {
		case "user":
			userOption = opt.UserValue(s)
		case "reason":
			reasonOption = opt.StringValue()
		}
	}

	if userOption == nil {
		respondWithError(s, i, "User param is missing or invalid.")
		return
	}

	if reasonOption == "" {
		respondWithError(s, i, "Reason param is missing or invalid.")
		return
	}

	moderator := i.Member
	if moderator == nil {
		respondWithError(s, i, "Could not get mod info.")
		return
	}

	member, err := s.GuildMember(i.GuildID, i.Member.User.ID)
	if err != nil {
		respondWithError(s, i, "Could not verify your permissions.")
		return
	}

	hasPermission := false

	for _, roleID := range member.Roles {
		role, err := s.State.Role(i.GuildID, roleID)
		if err != nil {
			continue
		}

		if (role.Permissions&discordgo.PermissionBanMembers) != 0 ||
			(role.Permissions&discordgo.PermissionAdministrator) != 0 {
			hasPermission = true
			break
		}
	}

	if !hasPermission {
		respondWithError(s, i, "You don't have permission to ban members!")
		return
	}

	protectedUsers := []string{
		"1190168282376966174",
		"1247213229076250710",
		"942494965219610717",
		"885995586959077386",
	}

	for _, id := range protectedUsers {
		if id == userOption.ID {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("%s cannot be banned by `%s`", userOption.Username, moderator.User.Username),
				},
			})
			return
		}
	}

	db := database.Get()

	var userAccount accounts.Users
	result := db.Where(`"discordId" = ?`, userOption.ID).First(&userAccount)
	if result.Error != nil {
		respondWithError(s, i, "This user doesn't have an account!")
		return
	}

	if userAccount.Banned {
		respondWithError(s, i, fmt.Sprintf("User `%s` is already banned!", userAccount.DisplayName))
		return
	}

	userAccount.Banned = true
	result = db.Model(&userAccount).Update("banned", true)
	if result.Error != nil {
		respondWithError(s, i, "An error occurred while processing the ban.")
		return
	}

	responseMessage := ""
	dmChannel, err := s.UserChannelCreate(userOption.ID)
	if err != nil {
		responseMessage = fmt.Sprintf("Banned `%s` but failed to send them a DM.", userAccount.DisplayName)
	} else {
		_, err = s.ChannelMessageSend(dmChannel.ID, fmt.Sprintf("You have been banned for reason: `%s`", reasonOption))
		if err != nil {
			responseMessage = fmt.Sprintf("Banned `%s` but failed to send them a DM.", userAccount.DisplayName)
		} else {
			responseMessage = fmt.Sprintf("Banned `%s` for reason: `%s`", userAccount.DisplayName, reasonOption)
		}
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: responseMessage,
		},
	})
}

func respondWithError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

package discord

import (
	"fmt"
	"log"
	"solana/discord/commands"
	"solana/utilities"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
)

func RegisterCommands(s *discordgo.Session) []*discordgo.ApplicationCommand {
	registeredCommands := make([]*discordgo.ApplicationCommand, 0)

	banCmd := commands.NewBanCommand()
	cmd, err := s.ApplicationCommandCreate(s.State.User.ID, "", &discordgo.ApplicationCommand{
		Name:        banCmd.Name,
		Description: banCmd.Description,
		Options:     banCmd.Options,
	})
	if err != nil {
		log.Printf("Error creating '%s' command: %v", banCmd.Name, err)
	} else {
		registeredCommands = append(registeredCommands, cmd)
		utilities.LogWithTimestamp(color.GreenString, "Registered command: "+banCmd.Name)
	}
	
	
	banMMCmd := commands.NewMatchmakingBanCommand()
	mmCmd, err := s.ApplicationCommandCreate(s.State.User.ID, "", &discordgo.ApplicationCommand{
		Name:        banMMCmd.Name,
		Description: banMMCmd.Description,
		Options:     banMMCmd.Options,
	})
	if err != nil {
		log.Printf("Error creating '%s' command: %v", banMMCmd.Name, err)
	} else {
		registeredCommands = append(registeredCommands, mmCmd)
		utilities.LogWithTimestamp(color.GreenString, "Registered command: "+banMMCmd.Name)
	}

	return registeredCommands
}

func SetupCommandHandlers(s *discordgo.Session) {
	banCmd := commands.NewBanCommand()
	banMMCmd := commands.NewMatchmakingBanCommand()
	
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionApplicationCommand {
			return
		}
		
		switch i.ApplicationCommandData().Name {
		case banCmd.Name:
			banCmd.Handle(s, i)
		case banMMCmd.Name:
			banMMCmd.Handle(s, i)
		}
	})
}
	


func Run() {
	discord, err := discordgo.New("Bot " + utilities.Get[string]("bot"))
	if err != nil {
		log.Fatal("Error: ", err)
	}

	discord.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		err := s.UpdateStatusComplex(discordgo.UpdateStatusData{
			Activities: []*discordgo.Activity{
				{
					Name: "Season X",
					Type: discordgo.ActivityTypeWatching,
				},
			},
			Status: string(discordgo.StatusOnline),
		})

		if err != nil {
			log.Println("Error setting activity:", err)
		}

		RegisterCommands(s)
		SetupCommandHandlers(s)

		utilities.LogWithTimestamp(color.GreenString, fmt.Sprintf("Logged in as: %s#%s", s.State.User.Username, s.State.User.Discriminator))
	})

	err = discord.Open()
	if err != nil {
		log.Fatal("Error: ", err)
	}

	utilities.LogWithTimestamp(color.GreenString, "Solana discord bot is now running!")
}

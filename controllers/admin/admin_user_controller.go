package admin

import (
	"encoding/json"
	"log"
	"net/http"
	"solana/modules/database"
	"solana/modules/database/entities/accounts"
	"solana/modules/database/entities/tournament"
	"solana/utilities"

	"github.com/gin-gonic/gin"
)

type AdminGetUserResponse struct {
	User              accounts.Users `json:"user"`
	FavoriteCharacter string         `json:"favoriteCharacter"`
	Arena             int            `json:"arena"`
}

func GetUser(c *gin.Context) {
	username := c.Param("username")
	log.Printf("Admin requesting user info for: %s", username)
	db := database.Get()
	var user accounts.Users
	if err := db.Where("\"displayName\" = ?", username).First(&user).Error; err != nil {
		log.Printf("Error finding user by displayName: %v", err)
		utilities.Internal.ServerError().Apply(c.Writer)
		return
	}
	log.Printf("Found user with AccountID: %s", user.AccountID)
	var tournamentScore tournament.TournamentScores
	tournamentErr := db.Where("account_id = ? AND type = ?", user.AccountID, "Hype_Lategame").First(&tournamentScore).Error
	if tournamentErr != nil {

		tournamentScore.Value = 0
	}

	var characterItem accounts.Items
	characterErr := db.Where("account_id = ? AND profile_id = ? AND template_id = ?",
		user.AccountID, "athena", "favorite_character").First(&characterItem).Error

	favoriteCharacter := "default"
	if characterErr != nil {
	} else {
		err := json.Unmarshal([]byte(characterItem.Value), &favoriteCharacter)
		if err != nil {
			log.Printf("Warning: Error unmarshaling favorite character: %v", err)
		}
	}

	response := AdminGetUserResponse{
		User:              user,
		FavoriteCharacter: favoriteCharacter,
		Arena:             tournamentScore.Value,
	}

	c.JSON(http.StatusOK, response)
}

package admin

import (
	"fmt"
	"math/rand"
	"solana/classes/mcp"
	"solana/managers"
	"solana/modules/database"
	"solana/modules/database/entities/accounts"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func Register(c *gin.Context) {
	var request struct {
		Email     string `json:"email" binding:"required"`
		Password  string `json:"password" binding:"required"`
		Username  string `json:"username" binding:"required"`
		Server    bool   `json:"server"`
		DiscordID string `json:"discord_id"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	accountId := uuid.New().String()
	accountId = strings.ReplaceAll(accountId, "-", "")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Failed to hash password",
		})
		return
	}

	discordId := fmt.Sprintf("%d", rand.Intn(899999999)+100000000)

	if request.DiscordID != "" {
		discordId = request.DiscordID
	}

	now := time.Now()

	newUser := accounts.Users{
		Email:                   request.Email,
		DisplayName:             request.Username,
		Password:                string(hashedPassword),
		AccountID:               accountId,
		DiscordID:               discordId,
		Role:                    "-1",
		Created:                 now,
		Banned:                  false,
		HasFL:                   false,
		BanHistory:              []map[string]string{},
		IsServer:                request.Server,
		PartyIndex:              0,
		LastLoginTime:           time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
		LastLoginIP:             "",
		HWID:                    "",
		ProfilePicture:          "https://cdn.discordapp.com/avatars//",
		DisplayNameChanges:      0,
		MatchmakingBannedUntil:  "",
		MatchmakingBannedSince:  "",
		MatchmakingBannedReason: "",
	}
	newScret := accounts.Secrets{
		Secret:    strings.ReplaceAll(uuid.New().String(), "-", ""),
		AccountID: accountId,
	}

	db := database.Get()
	tx := db.Begin()
	if tx.Error != nil {
		c.JSON(500, gin.H{
			"error": "Failed to start transaction",
		})
		return
	}

	if err := tx.Create(&newUser).Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{
			"error":   "Failed to create user",
			"details": err.Error(),
		})
		return
	}

	if err := tx.Create(&newScret).Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{
			"error":   "Failed to create secret",
			"details": err.Error(),
		})
		return
	}

	_, err = managers.CreateProfile(mcp.Athena, accountId, newUser.DisplayName)
	if err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{
			"error":   "Failed to create Athena profile",
			"details": err.Error(),
		})
		return
	}

	_, err = managers.CreateProfile(mcp.CommonCore, accountId, newUser.DisplayName)
	if err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{
			"error":   "Failed to create CommonCore profile",
			"details": err.Error(),
		})
		return
	}

	_, err = managers.CreateProfile(mcp.Creative, accountId, newUser.DisplayName)
	if err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{
			"error":   "Failed to create Creative profile",
			"details": err.Error(),
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(500, gin.H{
			"error":   "Failed to commit transaction",
			"details": err.Error(),
		})
		return
	}

	c.JSON(200, newScret)
}

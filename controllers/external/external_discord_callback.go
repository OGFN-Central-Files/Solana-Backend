package external

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"solana/classes/mcp"
	"solana/managers"
	"solana/modules/database"
	"solana/modules/database/entities/accounts"
	"solana/utilities"
	"solana/utilities/oauth"
)

type TokenResponse struct {
	AccessToken string `json:"access_token"`
}

type UserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	Verified bool   `json:"verified"`
}

func DiscordOAuthCallback(c *gin.Context) {
	db := database.Get()

	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid request V1"})
		return
	}

	client := &http.Client{}

	formData := url.Values{
		"client_id":     {utilities.Get[string]("client_id")},
		"client_secret": {utilities.Get[string]("client_secret")},
		"code":          {code},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {"http://127.0.0.1:5353/s/api/oauth/discord/callback"},
		"scope":         {"identify+email+guilds"},
	}

	tokenReq, err := http.NewRequest("POST", "https://discord.com/api/oauth2/token", strings.NewReader(formData.Encode()))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token request"})
		return
	}
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	tokenResp, err := client.Do(tokenReq)
	if err != nil || tokenResp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid request V2"})
		return
	}
	defer tokenResp.Body.Close()

	tokenBody, err := io.ReadAll(tokenResp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read token response"})
		return
	}

	var tokenData TokenResponse
	if err := json.Unmarshal(tokenBody, &tokenData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse token data"})
		return
	}

	userReq, err := http.NewRequest("GET", "https://discord.com/api/v10/users/@me", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user request"})
		return
	}
	userReq.Header.Set("Authorization", "Bearer "+tokenData.AccessToken)

	userResp, err := client.Do(userReq)
	if err != nil || userResp.StatusCode != http.StatusOK {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid User"})
		return
	}
	defer userResp.Body.Close()

	userBody, err := io.ReadAll(userResp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read user response"})
		return
	}

	var userData UserResponse
	if err := json.Unmarshal(userBody, &userData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user data"})
		return
	}

	if userData.ID == "" || userData.Username == "" || userData.Email == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid request V3"})
		return
	}

	var user accounts.Users
	result := db.Where("\"discordId\" = ?", userData.ID).First(&user)

	if result.Error != nil {
		accountID := strings.ReplaceAll(uuid.New().String(), "-", "")

		usernameHash := sha256.Sum256([]byte(userData.Username))
		pwHash := hex.EncodeToString(usernameHash[:])[:6]

		emailHash := sha256.Sum256([]byte(userData.Email))
		emailHashStr := hex.EncodeToString(emailHash[:])[:7]
		emailInt, _ := strconv.ParseInt(emailHashStr, 16, 64)
		encryptedEmail := fmt.Sprintf("%07x", emailInt) + "@solarisfn.org"

		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(pwHash), bcrypt.DefaultCost)

		tx := db.Begin()
		if tx.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
			return
		}

		newUser := accounts.Users{
			Email:                   encryptedEmail,
			DisplayName:             userData.Username,
			Password:                string(hashedPassword),
			AccountID:               accountID,
			DiscordID:               userData.ID,
			Role:                    strconv.Itoa(0),
			Created:                 time.Now(),
			Banned:                  false,
			HasFL:                   false,
			HasDiamond:              false,
			HasGold:                 false,
			HasPlat:                 false,
			BanHistory:              []map[string]string{},
			IsServer:                false,
			PartyIndex:              0,
			LastLoginTime:           time.Now().Format(time.RFC3339),
			LastLoginIP:             "",
			HWID:                    "",
			ProfilePicture:          fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s", userData.ID, userData.Avatar),
			DisplayNameChanges:      0,
			MatchmakingBannedUntil:  "",
			MatchmakingBannedSince:  "",
			MatchmakingBannedReason: "",
		}

		if err := tx.Create(&newUser).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		solosStats := map[string]int{
			"kills":         0,
			"matchesplayed": 0,
			"wins":          0,
			"top10":         0,
			"top25":         0,
		}
		duosStats := map[string]int{
			"kills":         0,
			"matchesplayed": 0,
			"wins":          0,
			"top5":          0,
			"top12":         0,
		}
		squadsStats := map[string]int{
			"kills":         0,
			"matchesplayed": 0,
			"wins":          0,
			"top3":          0,
			"top6":          0,
		}

		solosJSON, _ := json.Marshal(solosStats)
		duosJSON, _ := json.Marshal(duosStats)
		squadsJSON, _ := json.Marshal(squadsStats)

		newStats := accounts.Stats{
			AccountID:     accountID,
			Solos:         string(solosJSON),
			Duos:          string(duosJSON),
			Squads:        string(squadsJSON),
			MatchesPlayed: 0,
		}

		if err := tx.Create(&newStats).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create stats"})
			return
		}

		_, err := managers.CreateProfile(mcp.Athena, accountID, newUser.DisplayName)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Athena profile"})
			return
		}

		_, err = managers.CreateProfile(mcp.CommonCore, accountID, newUser.DisplayName)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create CommonCore profile"})
			return
		}

		_, err = managers.CreateProfile(mcp.Creative, accountID, newUser.DisplayName)
		if err != nil {
			tx.Rollback()
			c.JSON(500, gin.H{
				"error":   "Failed to create Creative profile",
				"details": err.Error(),
			})
			return
		}

		if err := tx.Commit().Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
			return
		}

		user = newUser
	}

	if err := db.Where("account_id = ? AND type = ?", user.AccountID, "exchange").Delete(&accounts.Tokens{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete existing tokens"})
		return
	}

	token, err := oauth.CreateExchangeToken(userData.ID, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
		return
	}

	user.ProfilePicture = fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s", userData.ID, userData.Avatar)

	if err := db.Model(&user).Updates(map[string]interface{}{
		"ProfilePicture": user.ProfilePicture,
		"Role":           user.Role,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, utilities.GenerateSuccessHTML(userData.ID, userData.Avatar, userData.Username, user.AccountID, token, user.DisplayName))
}

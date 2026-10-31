package oauth

import (
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"solana/modules/database"
	"solana/modules/database/entities/accounts"
	utilities "solana/utilities"
	oauth "solana/utilities/oauth"
)

func Create(c *gin.Context) {
	db := database.Get()
	tokenHeader := c.GetHeader("Authorization")

	if tokenHeader == "" {
		utilities.Authentication.InvalidHeader().Apply(c.Writer)
		return
	}

	client := strings.Split(tokenHeader, " ")
	if len(client) != 2 {
		utilities.Authentication.OAuth.InvalidClient().Apply(c.Writer)
		return
	}
	err := c.Request.ParseForm()
	if err != nil {
		utilities.Authentication.InvalidRequest().Apply(c.Writer)
		return
	}

	body := make(map[string]string)
	for key, values := range c.Request.PostForm {
		if key != "" {
			body[key] = values[0]
		}
	}

	grantType := body["grant_type"]
	if grantType == "" {
		utilities.Authentication.OAuth.UnsupportedGrant().Apply(c.Writer)
		return
	}

	clientID, err := base64.StdEncoding.DecodeString(client[1])
	if err != nil || string(clientID) == "" {
		utilities.Authentication.OAuth.InvalidClient().Apply(c.Writer)
		return
	}

	var user accounts.Users
	switch grantType {
	// case "exchange_code":
	// 	code := body["exchange_code"]

	// 	if code == "" {
	// 		utilities.Authentication.OAuth.InvalidAccountCredentials().Apply(c.Writer)
	// 		return
	// 	}

	// 	var exchangeCode accounts.ExchangeCodes
	// 	if err := db.Where("code = ?", code).First(&exchangeCode).Error; err != nil {
	// 		c.JSON(http.StatusBadRequest, utilities.Authentication.OAuth.InvalidAccountCredentials.Apply(c))
	// 		return
	// 	}

	// 	if err := db.Where("account_id = ?", exchangeCode.AccountID).First(&user).Error; err != nil {
	// 		c.JSON(http.StatusBadRequest, utilities.Authentication.OAuth.InvalidAccountCredentials.Apply(c))
	// 		return
	// 	}

	// 	db.Delete(&exchangeCode)

	case "device_auth":
		accountID := body["account_id"]

		if accountID == "" {
			utilities.Authentication.OAuth.InvalidAccountCredentials().Apply(c.Writer)
			return
		}

		if err := db.Where("account_id = ?", accountID).First(&user).Error; err != nil {
			utilities.Authentication.OAuth.InvalidAccountCredentials().Apply(c.Writer)
			return
		}

	case "password":
		username := body["username"]
		password := body["password"]

		if username == "" || password == "" {
			utilities.Authentication.OAuth.InvalidAccountCredentials().Apply(c.Writer)
			return
		}

		if err := db.Where("email = ?", username).First(&user).Error; err != nil {
			utilities.Authentication.OAuth.InvalidAccountCredentials().Apply(c.Writer)
			return
		}

		if user.Banned {
			utilities.Account.DisabledAccount().Apply(c.Writer)
			return
		}

		// err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		// if err != nil {
		// 	hash := helpers.GetDisplayNameHash(user.DisplayName)
		// 	if hash != password {
		// 		utilities.Authentication.OAuth.InvalidAccountCredentials().Apply(c.Writer)
		// 		return
		// 	}
		// }

	case "client_credentials":
		token, err := oauth.GenerateToken(string(clientID), string(clientID))
		if err != nil {
			utilities.Internal.ServerError().Apply(c.Writer)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"access_token":    token,
			"expires_in":      3600,
			"expires_at":      time.Now().Add(time.Hour).Format(time.RFC3339),
			"token_type":      "bearer",
			"client_id":       string(clientID),
			"internal_client": true,
			"client_service":  "fortnite",
		})
		return

	case "refresh_token":
		refreshToken := body["refresh_token"]
		if refreshToken == "" {
			utilities.Authentication.OAuth.InvalidRefresh().Apply(c.Writer)
			return
		}

		token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
			return []byte(utilities.Get[string]("jwt")), nil
		})
		if err != nil {
			utilities.Authentication.InvalidToken().Apply(c.Writer)
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		accountID := claims["sub"].(string)

		if err := db.Where("account_id = ?", accountID).First(&user).Error; err != nil {
			utilities.Account.AccountNotFound().Apply(c.Writer)
			return
		}

		if user.Banned {
			utilities.Account.DisabledAccount().Apply(c.Writer)
			return
		}

		accessToken, err := oauth.CreateAccessToken(string(clientID), grantType, user)
		if err != nil {
			utilities.Internal.ServerError().Apply(c.Writer)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"access_token":       accessToken,
			"expires_in":         3600,
			"expires_at":         time.Now().Add(time.Hour).Format(time.RFC3339),
			"token_type":         "bearer",
			"refresh_token":      body["refresh_token"],
			"refresh_expires":    86400,
			"refresh_expires_at": time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			"account_id":         user.AccountID,
			"client_id":          string(clientID),
			"internal_client":    true,
			"client_service":     "fortnite",
			"display_name":       user.DisplayName,
			"app":                "fortnite",
			"in_app_id":          user.AccountID,
			"device_id":          c.GetHeader("X-Epic-Device-Id"),
		})
		return

	default:
		utilities.Authentication.OAuth.GrantNotImplemented().Apply(c.Writer)
		return
	}

	if user.AccountID == "" {
		utilities.Authentication.OAuth.InvalidAccountCredentials().Apply(c.Writer)
		return
	}

	if !user.IsServer {
		db.Exec("DELETE FROM tokens WHERE account_id = ? AND type = ?", user.AccountID, "access")
		db.Exec("DELETE FROM tokens WHERE account_id = ? AND type = ?", user.AccountID, "refresh")
	}

	accessToken, err := oauth.CreateAccessToken(string(clientID), grantType, user)
	if err != nil {
		utilities.Internal.ServerError().Apply(c.Writer)
		return
	}

	refreshToken, err := oauth.CreateAccessToken(string(clientID), grantType, user)
	if err != nil {
		utilities.Internal.ServerError().Apply(c.Writer)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":       accessToken,
		"expires_in":         3600,
		"expires_at":         time.Now().Add(time.Hour).Format(time.RFC3339),
		"token_type":         "bearer",
		"account_id":         user.AccountID,
		"client_id":          string(clientID),
		"internal_client":    true,
		"client_service":     "fortnite",
		"refresh_token":      refreshToken,
		"refresh_expires":    86400,
		"refresh_expires_at": time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		"display_name":       user.DisplayName,
		"app":                "fortnite",
		"in_app_id":          user.AccountID,
		"device_id":          c.GetHeader("X-Epic-Device-Id"),
	})
}

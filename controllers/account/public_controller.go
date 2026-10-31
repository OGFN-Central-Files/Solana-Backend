package account

import (
	"net/http"

	"solana/modules/database"
	"solana/modules/database/entities/accounts"
	utilities "solana/utilities"

	"github.com/gin-gonic/gin"
)

func Normal(c *gin.Context) {
	accountID := c.Param("accountId")

	db := database.Get()
	var user accounts.Users
	if err := db.Where("\"accountId\" = ?", accountID).First(&user).Error; err != nil {
		utilities.Account.AccountNotFound().Apply(c.Writer)
		return
	}

	if user.Banned {
		utilities.Account.DisabledAccount().Apply(c.Writer)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":            user.AccountID,
		"displayName":   user.DisplayName,
		"externalAuths": gin.H{},
	})
}

func NormalExternal(c *gin.Context) {
	accountID := c.Param("accountId")

	db := database.Get()
	var user accounts.Users
	if err := db.Where("\"accountId\" = ?", accountID).First(&user).Error; err != nil {
		utilities.Account.AccountNotFound().Apply(c.Writer)
		return
	}

	if user.Banned {
		utilities.Account.DisabledAccount().Apply(c.Writer)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":            user.AccountID,
		"displayName":   user.DisplayName,
		"externalAuths": gin.H{},
	})
}

func Query(c *gin.Context) {
	accountIDs := c.QueryArray("accountId")

	if len(accountIDs) == 0 {
		utilities.Account.InvalidAccountIdCount().Apply(c.Writer)
		return
	}

	db := database.Get()
	var users []accounts.Users
	if err := db.Where("\"accountId\" IN ?", accountIDs).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query accounts"})
		return
	}

	response := make([]gin.H, 0)
	for _, user := range users {
		response = append(response, gin.H{
			"id":           user.AccountID,
			"displayName":  user.DisplayName,
			"externalAuth": gin.H{},
		})
	}

	c.JSON(http.StatusOK, response)
}

func DisplayName(c *gin.Context) {
	username := c.Param("username")

	db := database.Get()
	var user accounts.Users
	if err := db.Where("\"displayName\" = ?", username).First(&user).Error; err != nil {
		utilities.Account.AccountNotFound().Apply(c.Writer)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":            user.AccountID,
		"displayName":   user.DisplayName,
		"externalAuths": gin.H{},
	})
}

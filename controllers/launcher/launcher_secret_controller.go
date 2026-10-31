package launcher

import (
	"net/http"

	"solana/modules/database"
	"solana/modules/database/entities/accounts"
	utilities "solana/utilities"

	"github.com/gin-gonic/gin"

	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func QueryBySecret(c *gin.Context) {
	secret := c.Param("secret")

	db := database.Get()
	var s accounts.Secrets
	if err := db.Where("secret = ?", secret).First(&s).Error; err != nil {
		utilities.Account.AccountNotFound().Apply(c.Writer)
		return
	}

	var user accounts.Users
	if err := db.Where("\"accountId\" = ?", s.AccountID).First(&user).Error; err != nil {
		utilities.Account.AccountNotFound().Apply(c.Writer)
		return
	}

	hashedPassword := sha256.Sum256([]byte(user.DisplayName))
	pwhash := strings.ToLower(hex.EncodeToString(hashedPassword[:]))[:6]

	c.JSON(http.StatusOK, gin.H{
		"Secret":      s.Secret,
		"AccountId":   s.AccountID,
		"Email":       user.Email,
		"DisplayName": user.DisplayName,
		"Password":    pwhash,
	})
}

package oauth

import (
	"solana/modules/database"
	"solana/modules/database/entities/accounts"
	"solana/utilities"
	"strings"

	"github.com/gin-gonic/gin"
)

func ClientMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenHeader := c.GetHeader("Authorization")

		if tokenHeader == "" {
			utilities.Authentication.InvalidHeader().Apply(c.Writer)
			c.Abort()
			return
		}

		token := strings.ReplaceAll(tokenHeader, "bearer ", "")

		db := database.Get()
		var t accounts.Tokens
		if err := db.Where("token = ?", token).First(&t).Error; err != nil {
			utilities.Authentication.AuthenticationFailed().Apply(c.Writer)
			c.Abort()
			return
		}

		var user accounts.Users
		if err := db.Where("\"accountId\" = ?", t.AccountID).First(&user).Error; err != nil {
			utilities.Authentication.AuthenticationFailed().Apply(c.Writer)
			c.Abort()
			return
		}

		c.Next()
	}
}

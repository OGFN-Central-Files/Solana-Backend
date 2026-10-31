package synapse

import (
	"net/http"
	"solana/modules/database"
	"solana/modules/database/entities/accounts"
	"solana/utilities"

	"github.com/gin-gonic/gin"
)

func Auth(c *gin.Context) {
	id := c.Param("id")
	token := c.GetHeader("x-synapse")

	if token == "" || token != utilities.Get[string]("s_apitoken") {
		utilities.Account.DisabledAccount().Apply(c.Writer)
		return
	}

	db := database.Get()
	var user accounts.Users
	if err := db.Where("\"accountId\" = ?", id).First(&user).Error; err != nil {
		utilities.Account.AccountNotFound().Apply(c.Writer)
		return
	}

	c.Status(http.StatusNoContent)
}

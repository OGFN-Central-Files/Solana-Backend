package synapse

import (
	"net/http"
	"solana/modules/database"
	"solana/modules/database/entities/accounts"
	"solana/utilities"
	"strings"

	"github.com/gin-gonic/gin"
)


func Friends(c *gin.Context) {

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

	var friends []accounts.Friends
	if err := db.Where("account_id = ?", id).Find(&friends).Error; err != nil {
		utilities.Internal.ServerError().Apply(c.Writer)
		return
	}

	acceptedFriends := []string{}
	for _, friend := range friends {
		if strings.ToUpper(friend.Status) == "ACCEPTED" {
			acceptedFriends = append(acceptedFriends, friend.FriendId)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": acceptedFriends,
	})
}

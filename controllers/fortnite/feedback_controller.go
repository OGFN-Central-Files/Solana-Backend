package fortnite

import (
	"bytes"
	"encoding/json"
	"net/http"
	"solana/modules/database"
	"solana/modules/database/entities/accounts"
	utilities "solana/utilities"

	"github.com/gin-gonic/gin"
)

type FeedbackRequest struct {
	FeedbackType string `form:"feedbacktype"`
	AccountID    string `form:"accountid"`
	Subject      string `form:"subject"`
	FeedbackBody string `form:"feedbackbody"`
}

func SubmitBugFeedback(c *gin.Context) {
	var feedback FeedbackRequest
	if err := c.ShouldBind(&feedback); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feedback data"})
		return
	}

	db := database.Get()
	var user accounts.Users
	if err := db.Where("\"accountId\" = ?", feedback.AccountID).First(&user).Error; err != nil {
		utilities.Account.AccountNotFound().Apply(c.Writer)
		return
	}

	webhookURL := "https://discord.com/api/webhooks/1327780600777801820/RBWng24cDYL8-sU_TPXoorTLjwkBWLBKQ0nvcv39eZeOcaSED8lX9Lg_TK1rHXJ-JW_H"
	
	embed := map[string]interface{}{
		"username": "Solaris Bug Reporter",
		"embeds": []map[string]interface{}{
			{
				"title": "Bug Report",
				"color": 16711935,
				"fields": []map[string]interface{}{
					{"name": "Feedback Type", "value": feedback.FeedbackType, "inline": true},
					{"name": "Username", "value": user.DisplayName, "inline": true},
					{"name": "Subject", "value": feedback.Subject},
					{"name": "Details", "value": feedback.FeedbackBody},
				},
			},
		},
	}

	jsonData, err := json.Marshal(embed)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process feedback"})
		return
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil || resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send"})
		return
	}
	defer resp.Body.Close()

	c.JSON(http.StatusOK, gin.H{"message": "Feedback sent"})
}

func SubmitCommentFeedback(c *gin.Context) {
	var feedback FeedbackRequest
	if err := c.ShouldBind(&feedback); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feedback data"})
		return
	}

	db := database.Get()
	var user accounts.Users
	if err := db.Where("\"accountId\" = ?", feedback.AccountID).First(&user).Error; err != nil {
		utilities.Account.AccountNotFound().Apply(c.Writer)
		return
	}

	webhookURL := "https://discord.com/api/webhooks/1327780600777801820/RBWng24cDYL8-sU_TPXoorTLjwkBWLBKQ0nvcv39eZeOcaSED8lX9Lg_TK1rHXJ-JW_H"
	
	embed := map[string]interface{}{
		"username": "Solaris Comment Feedback",
		"embeds": []map[string]interface{}{
			{
				"title": "Comment Feedback",
				"color": 16711935,
				"fields": []map[string]interface{}{
					{"name": "Feedback Type", "value": feedback.FeedbackType, "inline": true},
					{"name": "Username", "value": user.DisplayName, "inline": true},
					{"name": "Subject", "value": feedback.Subject},
					{"name": "Details", "value": feedback.FeedbackBody},
				},
			},
		},
	}

	jsonData, err := json.Marshal(embed)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process feedback"})
		return
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil || resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send"})
		return
	}
	defer resp.Body.Close()

	c.JSON(http.StatusOK, gin.H{"message": "Feedback sent"})
}
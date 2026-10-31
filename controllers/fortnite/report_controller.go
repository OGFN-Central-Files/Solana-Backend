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


type ReportRequest struct {
	Reason  string `json:"reason"`
	Details string `json:"details"`
}


func ReportPlayer(c *gin.Context) {
	
	unsafeReporter := c.Param("unsafeReporter")
	reportedPlayer := c.Param("reportedPlayer")

	var reportRequest ReportRequest
	if err := c.ShouldBindJSON(&reportRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report data"})
		return
	}

	db := database.Get()
	
	var reportedPlayerUser accounts.Users
	if err := db.Where("\"accountId\" = ?", reportedPlayer).First(&reportedPlayerUser).Error; err != nil {
		utilities.Account.AccountNotFound().Apply(c.Writer)
		return
	}
	
	var reporterPlayerUser accounts.Users
	if err := db.Where("\"accountId\" = ?", unsafeReporter).First(&reporterPlayerUser).Error; err != nil {
		utilities.Account.AccountNotFound().Apply(c.Writer)
		return
	}

	reason := reportRequest.Reason
	if reason == "" {
		reason = "No reason provided"
	}
	
	details := reportRequest.Details
	if details == "" {
		details = "No details provided"
	}


	webhookURL := "https://discord.com/api/webhooks/1327780600777801820/RBWng24cDYL8-sU_TPXoorTLjwkBWLBKQ0nvcv39eZeOcaSED8lX9Lg_TK1rHXJ-JW_H"
	
	embed := map[string]interface{}{
		"username": "Solaris Report System",
		"embeds": []map[string]interface{}{
			{
				"title": "Solaris Report",
				"color": 16711680,
				"fields": []map[string]interface{}{
					{"name": "Reporter", "value": reporterPlayerUser.DisplayName, "inline": true},
					{"name": "Reported Player", "value": reportedPlayerUser.DisplayName, "inline": true},
					{"name": "Reason", "value": reason},
					{"name": "Details", "value": details},
				},
			},
		},
	}

	jsonData, err := json.Marshal(embed)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process report"})
		return
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil || resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit report"})
		return
	}
	defer resp.Body.Close()

	c.JSON(http.StatusOK, gin.H{"message": "Report submitted successfully"})
}
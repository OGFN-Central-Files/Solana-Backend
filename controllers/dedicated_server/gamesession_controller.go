package dedicated_server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"solana/modules/database"
	"solana/modules/database/entities/fortnite"
	"solana/utilities"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func CreateSession(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		utilities.Basic.BadRequest().Apply(c.Writer)
		return
	}

	id := uuid.New().String()
	id = id[0:8] + id[9:13] + id[14:18] + id[19:23] + id[24:]

	var publicPlayers []string
	var privatePlayers []string

	if pubPlayers, ok := body["publicPlayers"].([]interface{}); ok {
		for _, player := range pubPlayers {
			if playerStr, ok := player.(string); ok {
				publicPlayers = append(publicPlayers, playerStr)
			}
		}
	}

	if privPlayers, ok := body["privatePlayers"].([]interface{}); ok {
		for _, player := range privPlayers {
			if playerStr, ok := player.(string); ok {
				privatePlayers = append(privatePlayers, playerStr)
			}
		}
	}

	var attributesStr string
	if attrs, ok := body["attributes"]; ok {
		attributesBytes, err := json.Marshal(attrs)
		if err == nil {
			attributesStr = string(attributesBytes)
		}
	}

	session := fortnite.MMSessions{
		SessionId:                       id,
		ServerAddress:                   c.ClientIP(),
		LastUpdated:                     time.Now().UTC().Format(time.RFC3339),
		PlaylistName:                    "",
		OwnerId:                         body["ownerId"].(string),
		OwnerName:                       body["ownerName"].(string),
		ServerName:                      body["serverName"].(string),
		MaxPublicPlayers:                int(body["maxPublicPlayers"].(float64)),
		MaxPrivatePlayers:               int(body["maxPrivatePlayers"].(float64)),
		ShouldAdvertise:                 body["shouldAdvertise"].(bool),
		AllowJoinInProgress:             body["allowJoinInProgress"].(bool),
		IsDedicated:                     body["isDedicated"].(bool),
		UsesStats:                       body["usesStats"].(bool),
		AllowInvites:                    body["allowInvites"].(bool),
		UsesPresence:                    body["usesPresence"].(bool),
		AllowJoinViaPresence:            body["allowJoinViaPresence"].(bool),
		AllowJoinViaPresenceFriendsOnly: body["allowJoinViaPresenceFriendsOnly"].(bool),
		BuildUniqueId:                   body["buildUniqueId"].(string),
		Attributes:                      attributesStr,
		ServerPort:                      int(body["serverPort"].(float64)),
		OpenPublicPlayers:               int(body["openPublicPlayers"].(float64)),
		OpenPrivatePlayers:              int(body["openPrivatePlayers"].(float64)),
		SortWeight:                      int(body["sortWeight"].(float64)),
		Started:                         body["started"].(bool),
		PublicPlayers:                   publicPlayers,
		PrivatePlayers:                  privatePlayers,
		Stopped:                         false,
	}

	db := database.Get()
	if err := db.Create(&session).Error; err != nil {
		utilities.Matchmaking.UnknownSession().Apply(c.Writer)
		return
	}

	attributes := make(map[string]interface{})
	if session.Attributes != "" {
		json.Unmarshal([]byte(session.Attributes), &attributes)
	}

	response := map[string]interface{}{
		"id":                              session.SessionId,
		"serverAddress":                   session.ServerAddress,
		"lastUpdated":                     session.LastUpdated,
		"ownerId":                         session.OwnerId,
		"ownerName":                       session.OwnerName,
		"serverName":                      session.ServerName,
		"maxPublicPlayers":                session.MaxPublicPlayers,
		"maxPrivatePlayers":               session.MaxPrivatePlayers,
		"shouldAdvertise":                 session.ShouldAdvertise,
		"allowJoinInProgress":             session.AllowJoinInProgress,
		"isDedicated":                     session.IsDedicated,
		"usesStats":                       session.UsesStats,
		"allowInvites":                    session.AllowInvites,
		"usesPresence":                    session.UsesPresence,
		"allowJoinViaPresence":            session.AllowJoinViaPresence,
		"allowJoinViaPresenceFriendsOnly": session.AllowJoinViaPresenceFriendsOnly,
		"buildUniqueId":                   session.BuildUniqueId,
		"attributes":                      attributes,
		"serverPort":                      session.ServerPort,
		"openPublicPlayers":               session.OpenPublicPlayers,
		"openPrivatePlayers":              session.OpenPrivatePlayers,
		"sortWeight":                      session.SortWeight,
		"started":                         session.Started,
		"publicPlayers":                   session.PublicPlayers,
		"privatePlayers":                  session.PrivatePlayers,
	}

	c.JSON(http.StatusOK, response)
}

func CreateMatchmakingServiceTicket(c *gin.Context) {
	sessionId := c.Param("sessionId")
	if sessionId == "" {
		utilities.Basic.BadRequest().Apply(c.Writer)
		return
	}

	bucketIds := c.Query("bucketIds")
	if bucketIds == "" {
		utilities.Basic.BadRequest().Apply(c.Writer)
		return
	}

	bucketId := strings.Split(bucketIds, ":")
	if len(bucketId) < 4 {
		utilities.Basic.BadRequest().Apply(c.Writer)
		return
	}

	db := database.Get()
	var session fortnite.MMSessions
	if err := db.Where("session_id = ?", sessionId).First(&session).Error; err != nil {
		utilities.Matchmaking.UnknownSession().Apply(c.Writer)
		return
	}

	finalBucketIds := strings.Join(bucketId[:4], ":")

	randomBytes128 := make([]byte, 128)
	randomBytes32 := make([]byte, 32)

	if _, err := rand.Read(randomBytes128); err != nil {
		return
	}

	if _, err := rand.Read(randomBytes32); err != nil {
		return
	}

	jti := hex.EncodeToString(randomBytes32)

	now := time.Now().UTC()
	expirationTime := now.Add(240 * time.Minute)

	ua := utilities.Parse(c.GetHeader("User-Agent"))

	payload := jwt.MapClaims{
		"bucketId":      finalBucketIds,
		"region":        bucketId[2],
		"version":       ua.Build,
		"buildUniqueId": bucketId[0],
		"exp":           expirationTime.Unix(),
		"iat":           now.Unix(),
		"jti":           jti,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	tokenString, err := token.SignedString([]byte(utilities.Get[string]("jwt")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create token",
		})
		return
	}

	response := gin.H{
		"serviceUrl": utilities.Get[string]("sesh_socket"),
		"ticketType": "Iris-Sessions",
		"payload":    sessionId,
		"signature":  tokenString,
	}

	c.JSON(http.StatusOK, response)
}

func SessionHeartbeat(c *gin.Context) {
	sessionId := c.Param("sessionId")
	if sessionId == "" {
		utilities.Basic.BadRequest().Apply(c.Writer)
		return
	}

	db := database.Get()
	var session fortnite.MMSessions
	if err := db.Where("session_id = ?", sessionId).First(&session).Error; err != nil {
		utilities.Matchmaking.UnknownSession().Apply(c.Writer)
		return
	}

	session.LastUpdated = time.Now().UTC().Format(time.RFC3339)
	if err := db.Save(&session).Error; err != nil {
		utilities.Matchmaking.UnknownSession().Apply(c.Writer)
		return
	}

	response := map[string]interface{}{
		"sessionId":                       session.SessionId,
		"serverAddress":                   session.ServerAddress,
		"lastUpdated":                     session.LastUpdated,
		"ownerId":                         session.OwnerId,
		"ownerName":                       session.OwnerName,
		"serverName":                      session.ServerName,
		"maxPublicPlayers":                session.MaxPublicPlayers,
		"maxPrivatePlayers":               session.MaxPrivatePlayers,
		"shouldAdvertise":                 session.ShouldAdvertise,
		"allowJoinInProgress":             session.AllowJoinInProgress,
		"isDedicated":                     session.IsDedicated,
		"usesStats":                       session.UsesStats,
		"allowInvites":                    session.AllowInvites,
		"usesPresence":                    session.UsesPresence,
		"allowJoinViaPresence":            session.AllowJoinViaPresence,
		"allowJoinViaPresenceFriendsOnly": session.AllowJoinViaPresenceFriendsOnly,
		"buildUniqueId":                   session.BuildUniqueId,
		"attributes":                      session.Attributes,
		"serverPort":                      session.ServerPort,
		"openPublicPlayers":               session.OpenPublicPlayers,
		"openPrivatePlayers":              session.OpenPrivatePlayers,
		"sortWeight":                      session.SortWeight,
		"started":                         session.Started,
		"publicPlayers":                   session.PublicPlayers,
		"privatePlayers":                  session.PrivatePlayers,
		"stopped":                         session.Stopped,
	}

	c.JSON(http.StatusOK, response)
}

package matchmaking

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"solana/modules/database"
	"solana/modules/database/entities/accounts"
	"solana/modules/database/entities/fortnite"
	"solana/utilities"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func CreateMatchmakingServiceTicket(c *gin.Context) {
	accountId := c.Param("accountId")
	bucketId := c.Query("bucketId")
	partyPlayerIds := c.Query("partyPlayerIds")
	fillTeam := c.DefaultQuery("player.option.fillTeam", "")
	TournamentId := c.DefaultQuery("player.option.tournamentId", "")
	TournamentWindowId := c.DefaultQuery("player.option.windowId", "")

	if accountId == "" || bucketId == "" {
		utilities.Internal.ValidationFailed().Apply(c.Writer)
		return
	}

	decodedBucketId, err := url.QueryUnescape(bucketId)
	if err != nil {
		utilities.Internal.ValidationFailed().Apply(c.Writer)
		return
	}

	decodedBucketIds := strings.Split(decodedBucketId, ":")

	if TournamentId != "" && TournamentWindowId != "" {
		utilities.Internal.ValidationFailed().Apply(c.Writer)
		return
	}

	db := database.Get()
	var user accounts.Users
	if err := db.Where("\"accountId\" = ?", accountId).First(&user).Error; err != nil {
		utilities.Account.AccountNotFound().Apply(c.Writer)
		return
	}

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
		"bucketId":       bucketId,
		"partyPlayerIds": partyPlayerIds,
		"fillTeam":       fillTeam,
		"playlist":       decodedBucketIds[3],
		"region":         decodedBucketIds[2],
		"version":        ua.Build,
		"buildUniqueId":  decodedBucketIds[0],
		"exp":            expirationTime.Unix(),
		"iat":            now.Unix(),
		"jti":            jti,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	tokenString, err := token.SignedString([]byte(utilities.Get[string]("jwt")))
	if err != nil {
		utilities.Internal.ServerError().Apply(c.Writer)
		return
	}

	data := strings.Split(tokenString, ".")

	payloadString, err := json.Marshal(payload)
	if err != nil {
		utilities.Internal.ServerError().Apply(c.Writer)
		return
	}

	p := fortnite.MMPayloads{
		AccountId: accountId,
		Data:      string(payloadString),
	}
	db.Create(&p)

	response := gin.H{
		"serviceUrl": utilities.Get[string]("matchmaker"),
		"ticketType": "Iris",
		"payload":    data[0] + "." + data[1],
		"signature":  data[2],
	}

	c.JSON(http.StatusOK, response)
}

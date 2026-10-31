package oauth

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"

	"solana/modules/database"
	"solana/modules/database/entities/accounts"
	"solana/utilities"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func GenerateToken(clientID, clientSecret string) (string, error) {
	randomBytes128 := make([]byte, 128)
	randomBytes32 := make([]byte, 32)

	_, err := rand.Read(randomBytes128)
	if err != nil {
		return "", err
	}

	_, err = rand.Read(randomBytes32)
	if err != nil {
		return "", err
	}

	p := base64.StdEncoding.EncodeToString(randomBytes128)
	jti := hex.EncodeToString(randomBytes32)

	now := time.Now().UTC()
	expirationTime := now.Add(240 * time.Minute)

	payload := jwt.MapClaims{
		"p":             p,
		"clsvc":         "fortnite",
		"t":             "s",
		"mver":          false,
		"clid":          clientID,
		"ic":            true,
		"exp":           expirationTime.Unix(),
		"am":            "client_credentials",
		"iat":           now.Unix(),
		"jti":           jti,
		"creation_date": now.Format(time.RFC3339),
		"expires_in":    1,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	signedToken, err := token.SignedString([]byte(clientSecret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func CreateAccessToken(clientID, grantType string, user accounts.Users) (string, error) {
	payload := jwt.MapClaims{
		"app":           "fortnite",
		"sub":           user.AccountID,
		"dvid":          generateRandomNumber(),
		"mver":          false,
		"clid":          clientID,
		"dn":            user.DisplayName,
		"am":            "access",
		"p":             base64.StdEncoding.EncodeToString([]byte(generateRandomUUID())),
		"iai":           user.AccountID,
		"sec":           1,
		"clsvc":         "fortnite",
		"t":             "s",
		"ic":            true,
		"jti":           generateRandomUUID(),
		"creation_date": time.Now().UTC().Format(time.RFC3339),
		"expires_in":    4 * 3600,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	signedToken, err := token.SignedString([]byte(utilities.Get[string]("jwt")))
	if err != nil {
		fmt.Printf("wow creating token record: %v\n", err)
		return "", err
	}

	db := database.Get()
	//	permissions := createPermissions(user.AccountID, "access")

	tokenRecord := accounts.Tokens{
		Type:        "access",
		AccountID:   user.AccountID,
		Token:       signedToken,
		Permissions: []string{},
	}

	if err := db.Create(&tokenRecord).Error; err != nil {
		fmt.Printf("Error creating token record: %v\n", err)
		return "", err
	}

	return signedToken, nil
}

func CreateAdminToken(id string, user accounts.Users) (string, error) {
	payload := jwt.MapClaims{
		"discordId": id,
		"app":       "solaris",
		"sub":       user.AccountID,
		"dvid":      generateRandomNumber(),
		"dn":        user.DisplayName,
		"iai":       user.AccountID,
		"am":        "admin",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	signedToken, err := token.SignedString([]byte(utilities.Get[string]("jwt")))
	if err != nil {
		return "", err
	}

	db := database.Get()
	tokenRecord := accounts.Tokens{
		Type:        "admin",
		AccountID:   user.AccountID,
		Token:       signedToken,
		Permissions: []string{},
	}

	if err := db.Create(&tokenRecord).Error; err != nil {
		return "", err
	}

	return signedToken, nil
}

func CreateExchangeToken(id string, user accounts.Users) (string, error) {
	payload := jwt.MapClaims{
		"discordId": id,
		"app":       "solaris",
		"sub":       user.AccountID,
		"dvid":      generateRandomNumber(),
		"dn":        user.DisplayName,
		"iai":       user.AccountID,
		"am":        "exchange",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	signedToken, err := token.SignedString([]byte(utilities.Get[string]("jwt")))
	if err != nil {
		return "", err
	}

	db := database.Get()
	tokenRecord := accounts.Tokens{
		Type:        "exchange",
		AccountID:   user.AccountID,
		Token:       signedToken,
		Permissions: []string{},
	}

	if err := db.Create(&tokenRecord).Error; err != nil {
		return "", err
	}

	return signedToken, nil
}

func CreateRefreshToken(clientID, grantType string, user accounts.Users) (string, error) {
	payload := jwt.MapClaims{
		"app":           "fortnite",
		"sub":           user.AccountID,
		"dvid":          generateRandomNumber(),
		"mver":          false,
		"clid":          clientID,
		"dn":            user.DisplayName,
		"am":            "refresh",
		"p":             base64.StdEncoding.EncodeToString([]byte(generateRandomUUID())),
		"iai":           user.AccountID,
		"sec":           1,
		"clsvc":         "fortnite",
		"t":             "s",
		"ic":            true,
		"jti":           generateRandomUUID(),
		"creation_date": time.Now().UTC().Format(time.RFC3339),
		"expires_in":    14 * 24 * 3600,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	signedToken, err := token.SignedString([]byte(utilities.Get[string]("jwt")))
	if err != nil {
		return "", err
	}

	db := database.Get()
	tokenRecord := accounts.Tokens{
		Type:        "refresh",
		AccountID:   user.AccountID,
		Token:       signedToken,
		Permissions: []string{},
	}

	if err := db.Create(&tokenRecord).Error; err != nil {
		return "", err
	}

	return signedToken, nil
}

func generateRandomNumber() int {
	return rand.Intn(1000000000)
}

func generateRandomUUID() string {
	return fmt.Sprintf("%v", uuid.New())
}

package utilities

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	DBConn               string
	JWTSecret            string
	Port                 string
	Prod                 bool
	DISCORD_BotToken     string
	Maintenance          bool
	MatchMaker           string
	SeshSocket           string
	CFRegion             string
	CFAccessID           string
	CFSecretKey          string
	CFEndpoint           string
	SPassword            string
	SBaseURL             string
	SAPIToken            string
	CURRENT_VERSION      int
	DISCORD_CLIENTID     string
	DISCORD_CLIENTSECRET string
	Current_QuestWeek    int
}

var config Config
var configMap map[string]interface{}

func Load() {
	config = Config{
		DBConn:               os.Getenv("CONNECTION_URL"),
		JWTSecret:            os.Getenv("JWTSECRET"),
		Port:                 os.Getenv("PORT"),
		DISCORD_BotToken:     os.Getenv("DISCORD_BOT_TOKEN"),
		MatchMaker:           os.Getenv("MATCHMAKER"),
		SeshSocket:           os.Getenv("MM_SESSION_SOCKET"),
		CFRegion:             os.Getenv("CF_REGION"),
		CFAccessID:           os.Getenv("CF_ACCESS_ID"),
		CFSecretKey:          os.Getenv("CF_SECRET_KEY"),
		CFEndpoint:           os.Getenv("CF_ENDPOINT"),
		SPassword:            os.Getenv("S_PASSWORD"),
		SBaseURL:             os.Getenv("S_BASEURL"),
		SAPIToken:            os.Getenv("S_APITOKEN"),
		DISCORD_CLIENTID:     os.Getenv("DISCORD_CLIENTID"),
		DISCORD_CLIENTSECRET: os.Getenv("DISCORD_CLIENTSECRET"),
	}

	config.Prod, _ = strconv.ParseBool(os.Getenv("PROD"))
	config.Maintenance, _ = strconv.ParseBool(os.Getenv("MAINTENANCE"))

	currentVersionStr := os.Getenv("CURRENT_VERSION")
	if currentVersionStr == "" {
		config.CURRENT_VERSION = 10
	} else {
		currentVersion, err := strconv.Atoi(currentVersionStr)
		if err != nil {
			log.Fatalf("Error parsing CURRENT_VERSION: %v", err)
		}
		config.CURRENT_VERSION = currentVersion
	}

	currentQuestWeekStr := os.Getenv("Current_QuestWeek")
	if currentQuestWeekStr == "" {
		config.Current_QuestWeek = 10
	} else {
		currentQuestWeek, err := strconv.Atoi(currentQuestWeekStr)
		if err != nil {
			log.Fatalf("Error parsing Current_QuestWeek: %v", err)
		}
		config.Current_QuestWeek = currentQuestWeek
	}

	configMap = map[string]interface{}{
		"db_conn":       config.DBConn,
		"jwt":           config.JWTSecret,
		"port":          config.Port,
		"prod":          config.Prod,
		"bot":           config.DISCORD_BotToken,
		"maintenance":   config.Maintenance,
		"matchmaker":    config.MatchMaker,
		"sesh_socket":   config.SeshSocket,
		"cf_region":     config.CFRegion,
		"cf_access_id":  config.CFAccessID,
		"cf_secret_key": config.CFSecretKey,
		"cf_endpoint":   config.CFEndpoint,
		"s_password":    config.SPassword,
		"s_baseurl":     config.SBaseURL,
		"s_apitoken":    config.SAPIToken,
		"client_id":     config.DISCORD_CLIENTID,
		"client_secret": config.DISCORD_CLIENTSECRET,
	}
}

func Get[T any](key string) T {
	var zero T

	val, exists := configMap[key]
	if !exists {
		return zero
	}

	typedVal, ok := val.(T)
	if ok {
		return typedVal
	}

	return zero
}

func GetConfig() Config {
	return config
}

package main

import (
	"log"
	"time"

	"solana/controllers/account"
	"solana/controllers/admin"
	"solana/controllers/cloudstorage"
	"solana/controllers/dedicated_server"
	"solana/controllers/external"
	"solana/controllers/fortnite"
	"solana/controllers/fortnite/creative"
	dsmcp "solana/controllers/fortnite/ds_mcp"
	"solana/controllers/fortnite/mcp"
	"solana/controllers/launcher"
	"solana/controllers/matchmaking"
	"solana/controllers/oauth"
	"solana/controllers/synapse"
	"solana/discord"
	"solana/managers"

	//"solana/discord"

	"solana/modules/database"
	accounts "solana/modules/database/entities/accounts"
	"solana/modules/database/entities/contentpages"
	fort "solana/modules/database/entities/fortnite"
	"solana/modules/database/entities/solaris"
	"solana/modules/database/entities/tournament"
	"solana/modules/storefront"
	"solana/utilities"
	auth "solana/utilities/oauth"

	"github.com/fatih/color"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	log.SetFlags(0)
	gin.SetMode(gin.ReleaseMode)

	if err := godotenv.Load(); err != nil {
		color.Yellow("Warning: Could not load .env file: %v", err)
	} else {
		utilities.Load()
	}

	db, err := database.Init()
	if err != nil {
		color.Red("Failed to connect to database: %v", err)
		log.Fatalf("Failed to connect to database: %v", err)
	}

	db.AutoMigrate(
		&accounts.Users{},
		&accounts.Tokens{},
		&accounts.Stats{},
		&accounts.Quests{},
		&accounts.Profiles{},
		&accounts.ExchangeCodes{},
		&accounts.Friends{},
		&accounts.Secrets{},
		&accounts.Items{},
		&fort.Hotfixes{},
		&fort.Catalog{},
		&fort.Battlepass{},
		&fort.CreatorCodes{},
		&fort.MMPayloads{},
		&fort.MMSessions{},
		&fort.MatchStats{},
		&fort.PastShops{},
		&contentpages.PlaylistInfo{},
		&contentpages.TournamentInfo{},
		&tournament.TournamentDedicatedPlayers{},
		&tournament.TournamentEvent{},
		&tournament.TournamentHistory{},
		&tournament.TournamentScores{},
		&tournament.TournamentTemplate{},
		&tournament.TournamentTokens{},
		&solaris.Banners{},
		&solaris.Files{},
		&solaris.Posts{},
		&solaris.Updater{},
		&fort.ShopSections{},
	)

	discord.Run()

	router := gin.New()
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "issuer", "Issuer", "authorization"},
	}))

	storefront.Init()
	cloudstorage.Init()
	router.Use(func(c *gin.Context) {
		startTime := time.Now()
		c.Next()

		if c.Request.URL.Path == "/images/icons/gear.png" || c.Request.URL.Path == "/favicon.ico" {
			return
		}

		duration := time.Since(startTime)

		statusColor := color.New(color.FgWhite).SprintFunc()
		statusCode := c.Writer.Status()

		switch {
		case statusCode >= 200 && statusCode < 300:
			statusColor = color.New(color.FgGreen).SprintFunc()
		case statusCode >= 300 && statusCode < 400:
			statusColor = color.New(color.FgCyan).SprintFunc()
		case statusCode >= 400 && statusCode < 500:
			statusColor = color.New(color.FgYellow).SprintFunc()
		case statusCode >= 500:
			statusColor = color.New(color.FgRed).SprintFunc()
		}

		utilities.LogWithTimestamp(color.WhiteString, "%s %s %s %.2fms",
			color.YellowString(c.Request.Method),
			color.CyanString(c.Request.URL.Path),
			statusColor(statusCode),
			float64(duration.Milliseconds())+float64(duration.Nanoseconds()%1e6)/1e6,
		)
	})

	router.Use(func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				utilities.LogWithTimestamp(color.RedString, "Error! %v", e.Err.Error())
			}

			if !c.Writer.Written() {
				utilities.Internal.ServerError().Apply(c.Writer)
			}
		}
	})

	go func() {
		managers.CacheMissionBundles()
		managers.CacheBattlePass()
		managers.CacheCosmeticVariantTokens()
		managers.CacheBattlePassStars()
		managers.CacheAthenaSeasonXPCurve()
		_, err = managers.CacheDailyQuests()
		if err != nil {
			utilities.LogWithTimestamp(color.RedString, "Error!: %v", err)
		}
		mcp.LoadCache()
	}()

	router.NoRoute(func(c *gin.Context) {
		utilities.Basic.NotFound().Apply(c.Writer)
	})

	router.HandleMethodNotAllowed = true
	router.NoMethod(func(c *gin.Context) {
		utilities.Basic.MethodNotAllowed().Apply(c.Writer)
	})

	router.POST("/s/api/server/register", admin.Register)
	router.GET("/s/api/admin/hotfixes", admin.GetHotfixes)
	router.GET("/s/api/admin/hotfixes/:file", admin.GetHotfix)
	router.POST("/s/api/admin/hotfixes/:file", admin.UpdateHotfix)
	router.GET("/s/api/admin/user/:username", admin.GetUser)

	router.POST("/account/api/oauth/token", oauth.Create)
	router.GET("/fortnite/api/v2/versioncheck/:platform", fortnite.VersionCheck)
	router.GET("/fortnite/api/cloudstorage/system", cloudstorage.GetCloudstorage)
	router.GET("/fortnite/api/cloudstorage/system/config", cloudstorage.GetCloudstorage)
	router.GET("/fortnite/api/cloudstorage/system/:filename", cloudstorage.GetCloudstorageFile)
	router.POST("/fortnite/api/game/v2/tryPlayOnPlatform/account/:accountId", fortnite.TryPlayOnPlatform)
	router.GET("/fortnite/api/game/v2/enabled_features", fortnite.EnabledFeatures)
	router.GET("/lightswitch/api/service/bulk/status", fortnite.LightswitchBulk)
	router.GET("/waitingroom/api/waitingroom", fortnite.Ret204)
	router.POST("/api/v1/assets/Fortnite/:version/:cl", fortnite.Ret204)
	router.GET("/fortnite/api/game/v2/privacy/account/:accountdId", fortnite.Ret204)
	router.POST("/datarouter/api/v1/public/data", fortnite.Ret204)
	router.GET("/content/api/pages/fortnite-game", fortnite.FortniteCPHandler)
	router.GET("/receipts/v1/account/:accountId/receipts", fortnite.Ret204)
	router.POST("/fortnite/api/game/v2/toxicity/account/:unsafeReporter/report/:reportedPlayer", fortnite.ReportPlayer)

	FeedBack := router.Group("/fortnite/api/feedback")
	{
		FeedBack.Use(auth.ClientMiddleware())
		FeedBack.POST("/Bug", fortnite.SubmitBugFeedback)
		FeedBack.POST("/Comment", fortnite.SubmitCommentFeedback)
	}

	CloudstorageUser := router.Group("/fortnite/api/cloudstorage/user")
	{
		CloudstorageUser.Use(auth.ClientMiddleware())
		CloudstorageUser.GET("/:accountId", cloudstorage.GetUsersCloudstorageFiles2)
		CloudstorageUser.GET("/:accountId/:filename", cloudstorage.GetUserCloudstorageFile)
		CloudstorageUser.PUT("/:accountId/:filename", cloudstorage.SaveUsersCloudstorageFile)
	}

	Calendar := router.Group("/fortnite/api/calendar/v1")
	{
		Calendar.Use(auth.ClientMiddleware())
		Calendar.GET("/timeline", fortnite.Timeline)
	}

	Storefront := router.Group("/fortnite/api/storefront/v2")
	{
		Storefront.Use(auth.ClientMiddleware())
		Storefront.GET("/catalog", fortnite.Catalog)
		Storefront.GET("/keychain", fortnite.Keychain)
	}

	AccountPublic := router.Group("/account/api/public/account")
	{
		AccountPublic.Use(auth.ClientMiddleware())
		AccountPublic.GET("/:accountId", account.Normal)
		AccountPublic.GET("/:accountId/externalAuths", account.NormalExternal)
		AccountPublic.GET("/displayName/:username", account.DisplayName)
		AccountPublic.GET("/", account.Query)
	}

	MCP := router.Group("/fortnite/api/game/v2/profile/:accountId")
	{
		MCP.Use(auth.ClientMiddleware())
		MCP.POST("/client/QueryProfile", mcp.QueryProfile)
		MCP.POST("/client/ClientQuestLogin", mcp.ClientQuestLogin)
		MCP.POST("/client/SetMtxPlatform", mcp.SetMtxPlatform)
		MCP.POST("/client/EquipBattleRoyaleCustomization", mcp.EquipBattleRoyaleCustomization)
		MCP.POST("/client/SetCosmeticLockerSlot", mcp.SetCosmeticLockerSlot)
		MCP.POST("/client/MarkItemSeen", mcp.MarkItemSeen)
		MCP.POST("/client/SetItemFavoriteStatusBatch", mcp.SetItemFavoriteStatusBatch)
		MCP.POST("/client/SetBattleRoyaleBanner", mcp.SetBattleRoyaleBanner)
		MCP.POST("/client/PurchaseCatalogEntry", mcp.PurchaseCatalogEntry)
		MCP.POST("/client/RemoveGiftBox", mcp.RemoveGiftBox)
		MCP.POST("/client/BulkEquipBattleRoyaleCustomization", mcp.QueryProfile)
		MCP.POST("/client/SetPartyAssistQuest", mcp.SetPartyAssistQuest)
		MCP.POST("/client/SetMatchmakingBansViewed", mcp.QueryProfile)
	}

	DS_MCP := router.Group("/fortnite/api/game/v2/profile/:accountId/dedicated_server")
	{
		DS_MCP.POST("/QueryProfile", mcp.QueryProfile)
		DS_MCP.POST("/CreateNewIsland", dsmcp.CreateNewIsland)
		DS_MCP.POST("/UpdateQuests", dsmcp.UpdateQuests)
		DS_MCP.POST("/SendPlayerMatchStats", dsmcp.SendPlayerMatchStats)
		DS_MCP.GET("/QueryCreativeIslands", dsmcp.QueryCreativeIslands)
		DS_MCP.GET("/QueryCreativeIsland/:uuid", dsmcp.QueryCreativeIsland)
		DS_MCP.POST("/SetCreativePlotMetadata", dsmcp.SetCreativePlotMetadata)
	}

	External := router.Group("/s/api/oauth/discord")
	{
		External.GET("/", external.DiscordOAuthRedirect)
		External.GET("/callback", external.DiscordOAuthCallback)
	}

	Launcher := router.Group("/s/api/v2/launcher")
	{
		Launcher.GET("/hosting/user/:secret", launcher.QueryBySecret)
		Launcher.GET("/posts", launcher.GetLauncherPostsV2)
		Launcher.GET("/files", launcher.GetLauncherFiles)
		Launcher.GET("/updater", launcher.GetLauncherUpdater)
		Launcher.GET("/trailer", launcher.GetLauncherTrailer)
		Launcher.GET("/news", launcher.GetLauncherNews)
		Launcher.GET("/download", launcher.GetLauncherUpdate)
		Launcher.GET("/account", launcher.GetAccountLauncher)
		Launcher.GET("/shop", launcher.GetStorefront)
		Launcher.GET("/", launcher.GetLauncher)
	}

	Synapse := router.Group("/s/api/synapse")
	{
		Synapse.POST("/:id/auth", synapse.Auth)
		Synapse.POST("/:id/friends", synapse.Friends)
	}

	Matchmaking := router.Group("/fortnite/api/matchmaking")
	{
		Matchmaking.POST("/session", dedicated_server.CreateSession)
		Matchmaking.POST("/session/:sessionId/heartbeat", dedicated_server.SessionHeartbeat)
	}

	MatchmakingService := router.Group("/fortnite/api/game/v2/matchmakingservice")
	{
		MatchmakingService.POST("/ticket/session/:sessionId", dedicated_server.CreateMatchmakingServiceTicket)
		MatchmakingService.GET("/ticket/player/:accountId", matchmaking.CreateMatchmakingServiceTicket)
		MatchmakingService.Use(auth.ClientMiddleware())
	}

	LinksService := router.Group("/links/api/fn")
	{
		LinksService.GET("/mnemonic/:id", creative.GetLinkViaMnemonic)
	}

	utilities.LogWithTimestamp(color.GreenString, "Solana server running on port %s", utilities.Get[string]("port"))
	port := utilities.Get[string]("port")
	address := "0.0.0.0:" + port
	if err := router.Run(address); err != nil {
		log.Fatalf("Failed to run solana: %v", err)
	}
}

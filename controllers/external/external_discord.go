package external

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

func DiscordOAuthRedirect(c *gin.Context) {
	clientID := "1229597606133497938"
	redirectURI := url.QueryEscape("http://127.0.0.1:5353/s/api/oauth/discord/callback")
	scope := "identify+email+guilds"

	redirectURL := fmt.Sprintf(
		"https://discord.com/api/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&prompt=none",
		clientID, redirectURI, scope,
	)

	c.Redirect(http.StatusFound, redirectURL)
}

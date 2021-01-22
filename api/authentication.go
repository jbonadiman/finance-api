package handler

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/jbonadiman/finance-bot/databases"
	"github.com/jbonadiman/finance-bot/utils"
	"golang.org/x/oauth2"
	"io"
	"net/http"
	"time"
)

var (
	cacheClient *redis.Client
)

func init() {
	cacheClient = databases.GetClient()
}

func LoginRedirect(w http.ResponseWriter, r *http.Request) {
	microsoftConsumerEndpoint := oauth2.Endpoint{}

	clientId := utils.LoadVarSendingResponse(&w, "MS_CLIENT_ID")
	clientSecret := utils.LoadVarSendingResponse(&w, "MS_CLIENT_SECRET")
	authRedirectUrl := utils.LoadVarSendingResponse(&w, "MS_REDIRECT")

	microsoftConsumerEndpoint.AuthStyle = oauth2.AuthStyleInHeader
	microsoftConsumerEndpoint.AuthURL = "https://login.microsoftonline.com/consumers/oauth2/v2.0/authorize"
	microsoftConsumerEndpoint.TokenURL = "https://login.microsoftonline.com/consumers/oauth2/v2.0/token"

	msConfig := &oauth2.Config{
		RedirectURL:  authRedirectUrl,
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Scopes:       []string{"offline_access tasks.readwrite"},
		Endpoint:     microsoftConsumerEndpoint,
	}

	query := r.URL.Query()

	token, err := msConfig.Exchange(
		context.Background(),
		query.Get("code"))
	if err != nil {
		utils.SendError(&w, err)
	}

	cacheClient.Set(
		context.Background(),
		"token",
		token.AccessToken,
		token.Expiry.Sub(time.Now()))

	io.WriteString(w, "Authentication was successful!")
}

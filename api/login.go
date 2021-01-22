package handler

import (
	"github.com/google/uuid"
	"github.com/jbonadiman/finance-bot/utils"
	"golang.org/x/oauth2"
	"net/http"
)

func Login(w http.ResponseWriter, r *http.Request) {
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

	url := msConfig.AuthCodeURL(uuid.New().String())
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

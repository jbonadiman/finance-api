package handler

import (
	"log"
	"net/http"

	"github.com/google/uuid"
	"golang.org/x/oauth2"

	"github.com/jbonadiman/finance-bot/environment"
)

var (
	MSConfig *oauth2.Config
	authState uuid.UUID
)

const (
	msAuthURL  = "https://login.microsoftonline.com/consumers/oauth2/v2.0/authorize"
	msTokenURL = "https://login.microsoftonline.com/consumers/oauth2/v2.0/token"
	tasksScope = "tasks.readwrite"
)

func init() {
	consumerEndpoint := oauth2.Endpoint{}

	consumerEndpoint.AuthStyle = oauth2.AuthStyleInHeader
	consumerEndpoint.AuthURL = msAuthURL
	consumerEndpoint.TokenURL = msTokenURL

	MSConfig = &oauth2.Config{
		RedirectURL:  environment.MSRedirectURL,
		ClientID:     environment.MSClientID,
		ClientSecret: environment.MSClientSecret,
		Scopes:       []string{tasksScope},
		Endpoint:     consumerEndpoint,
	}
}

func Index(w http.ResponseWriter, r *http.Request) {
	var url string

	log.Println("getting url of authorize endpoint...")
	url = MSConfig.AuthCodeURL(authState.String())

	log.Printf("redirecting to: %v...", url)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

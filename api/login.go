package handler

import (
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"io"
	"log"
	"net/http"
	"os"
)

func Login(w http.ResponseWriter, r *http.Request) {
	microsoftConsumerEndpoint := oauth2.Endpoint{}

	clientId := os.Getenv("MS_CLIENT_ID")
	if clientId == "" {
		log.Printf("%q environment variable must be set!", "MS_CLIENT_ID")

		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, fmt.Sprintf("%q environment variable must be set!", "MS_CLIENT_ID"))
	}

	clientSecret := os.Getenv("MS_CLIENT_SECRET")
	if clientSecret == "" {
		log.Printf("%q environment variable must be set!", "MS_CLIENT_SECRET")

		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, fmt.Sprintf("%q environment variable must be set!", "MS_CLIENT_SECRET"))
	}

	authRedirectUrl := os.Getenv("MS_REDIRECT")
	if authRedirectUrl == "" {
		log.Printf("%q environment variable must be set!", "MS_REDIRECT")

		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, fmt.Sprintf("%q environment variable must be set!", "MS_REDIRECT"))
	}

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

package handler

import (
	"github.com/google/uuid"
	"github.com/jbonadiman/finance-bot/databases/redis"
	"github.com/jbonadiman/finance-bot/utils"
	"golang.org/x/oauth2"
	"log"
	"net/http"
)

var (
	MSClientID     string
	MSClientSecret string
	MSRedirectUrl  string

	MSConsumerEndpoint oauth2.Endpoint
	MSConfig *oauth2.Config
)

func init() {
	var err error

	MSClientID, err = utils.LoadVar("MS_CLIENT_ID")
	if err != nil {
		log.Println(err.Error())
	}

	MSClientSecret, err = utils.LoadVar("MS_CLIENT_SECRET")
	if err != nil {
		log.Println(err.Error())
	}

	MSRedirectUrl, err = utils.LoadVar("MS_REDIRECT")
	if err != nil {
		log.Println(err.Error())
	}

	MSConsumerEndpoint := oauth2.Endpoint{}

	MSConsumerEndpoint.AuthStyle = oauth2.AuthStyleInHeader
	MSConsumerEndpoint.AuthURL = "https://login.microsoftonline.com/consumers/oauth2/v2.0/authorize"
	MSConsumerEndpoint.TokenURL = "https://login.microsoftonline.com/consumers/oauth2/v2.0/token"

	MSConfig = &oauth2.Config{
		RedirectURL:  MSRedirectUrl,
		ClientID:     MSClientID,
		ClientSecret: MSClientSecret,
		Scopes:       []string{"offline_access", "tasks.readwrite"},
		Endpoint:     MSConsumerEndpoint,
	}
}

func Index(w http.ResponseWriter, r *http.Request) {
	if MSClientID == "" || MSClientSecret == "" || MSRedirectUrl == "" {
		http.Error(w, "Microsoft credentials environment variables must be set", http.StatusBadRequest)
	}

	cachedToken, err := redis.GetTokenFromCache()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	var url string

	if cachedToken != "" {
		log.Println("retrieved token from cache...")
		url = "/get-tasks"
	} else {
		log.Println("getting url of authorize endpoint...")
		url = MSConfig.AuthCodeURL(uuid.New().String())
	}

	log.Printf("redirecting to: %v...", url)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

package handler

import (
	"log"
	"net/http"

	"github.com/google/uuid"
	"golang.org/x/oauth2"

	"github.com/jbonadiman/finance-bot/app_msgs"
	"github.com/jbonadiman/finance-bot/databases/redis"
	//	_ "github.com/jbonadiman/finance-bot/events/consumers"
	"github.com/jbonadiman/finance-bot/utils"
)

var (
	MSClientID     string
	MSClientSecret string
	MSRedirectUrl  string

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

	consumerEndpoint := oauth2.Endpoint{}

	consumerEndpoint.AuthStyle = oauth2.AuthStyleInHeader
	consumerEndpoint.AuthURL = "https://login.microsoftonline.com/consumers/oauth2/v2.0/authorize"
	consumerEndpoint.TokenURL = "https://login.microsoftonline.com/consumers/oauth2/v2.0/token"

	MSConfig = &oauth2.Config{
		RedirectURL:  MSRedirectUrl,
		ClientID:     MSClientID,
		ClientSecret: MSClientSecret,
		Scopes:       []string{"offline_access", "tasks.readwrite"},
		Endpoint:     consumerEndpoint,
	}
}

func Index(w http.ResponseWriter, r *http.Request) {
	log.Println("checking for microsoft credentials in environment variables...")
	if MSClientID == "" || MSClientSecret == "" || MSRedirectUrl == "" {
		app_msgs.SendBadRequest(&w, app_msgs.MsCredentials())
		return
	}

	cachedToken, err := redis.GetTokenFromCache()
	if err != nil {
		app_msgs.SendInternalError(&w, err.Error())
		return
	}

	var url string

	if cachedToken != "" {
		url = "/api/get-tasks"
	} else {
		log.Println("getting url of authorize endpoint...")
		url = MSConfig.AuthCodeURL(uuid.New().String())
	}

	log.Printf("redirecting to: %v...", url)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

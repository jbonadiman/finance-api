package workers

import (
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"

	"github.com/jbonadiman/finance-bot/environment"
)

var (
	MSConfig  *oauth2.Config
	AuthState uuid.UUID
	delay = time.Tick(environment.AuthCronDuration)
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

func RequestAuthPage() {
	for range delay {
		AuthState = uuid.New()

		url := MSConfig.AuthCodeURL(AuthState.String())

		log.Printf("requesting auth page: %v...", url)
		resp, err := http.Get(url)

		if err != nil {
			log.Printf("error executing auth request: %v\n", err.Error())
		}

		log.Println(resp.Location())
	}
}

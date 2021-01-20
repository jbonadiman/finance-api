package services

import (
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"os"
)

const (
	clientIdEnv     = "MS_CLIENT_ID"
	clientSecretEnv = "MS_CLIENT_SECRET"
	redirectUrl     = "http://localhost:8080/authentication"
)

var (
	clientId                  string
	clientSecret              string
	microsoftConsumerEndpoint oauth2.Endpoint
)

type AuthHandler interface {
	Login(http.ResponseWriter, *http.Request)
	LoginRedirect(http.ResponseWriter, *http.Request)
}

type AuthService struct {
	RedirectUrl         string
	authState           string
	microsoftAuthConfig *oauth2.Config
}

func init() {
	clientId = os.Getenv(clientIdEnv)
	if clientId == "" {
		log.Fatalf("%q environment variable must be set!", clientIdEnv)
	}

	clientSecret = os.Getenv(clientSecretEnv)
	if clientSecret == "" {
		log.Fatalf("%q environment variable must be set!", clientSecretEnv)
	}
}

func New() *AuthService {
	microsoftConsumerEndpoint.AuthStyle = oauth2.AuthStyleInHeader
	microsoftConsumerEndpoint.AuthURL = "https://login.microsoftonline.com/consumers/oauth2/v2.0/authorize"
	microsoftConsumerEndpoint.TokenURL = "https://login.microsoftonline.com/consumers/oauth2/v2.0/token"

	config := &oauth2.Config{
		RedirectURL:  redirectUrl,
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Scopes:       []string{"offline_access tasks.readwrite"},
		Endpoint:     microsoftConsumerEndpoint,
	}

	service := AuthService{
		RedirectUrl:         redirectUrl,
		authState:           "",
		microsoftAuthConfig: config,
	}

	return &service
}

func (auth *AuthService) Login(w http.ResponseWriter, r *http.Request) {
	oauthStateString = uuid.New().String()

	url := microsoftOAuthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (auth *AuthService) LoginRedirect(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	token, err := getAccessToken(query.Get("state"), query.Get("code"))
	if err != nil {
		fmt.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	fmt.Fprintf(w, "Access token: %s\n", token)
}

func getAccessToken(state string, code string) (*oauth2.Token, error) {
	if state != oauthStateString {
		return nil, fmt.Errorf("invalid oauth state")
	}

	token, err := microsoftOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}

	return token, nil
	//response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	//if err != nil {
	//	return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	//}
	//defer response.Body.Close()
	//contents, err := ioutil.ReadAll(response.Body)
	//if err != nil {
	//	return nil, fmt.Errorf("failed reading response body: %s", err.Error())
	//}
	//return contents, nil
}

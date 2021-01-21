package auth_service

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jbonadiman/finance-bot/src/services"
	"github.com/labstack/echo/v4"
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
	Login(echo.Context) error
	LoginRedirect(echo.Context) error
}

type AuthService struct {
	services *[]services.Authenticated
	RedirectUrl string
	Token       *oauth2.Token

	state               string
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

func New(auth *[]services.Authenticated) *AuthService {
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
		services:            auth,
		RedirectUrl:         redirectUrl,
		state:               "",
		microsoftAuthConfig: config,
	}

	return &service
}

func (auth *AuthService) Login(c echo.Context) error {
	auth.state = uuid.New().String()

	url := auth.microsoftAuthConfig.AuthCodeURL(auth.state)
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

func (auth *AuthService) LoginRedirect(c echo.Context) error {
	token, err := auth.getAccessToken(c.QueryParam("state"), c.QueryParam("code"))
	if err != nil {
		c.Logger().Error(err.Error())
		return c.Redirect(http.StatusTemporaryRedirect, "/")
	}

	for i := 0; i < len(*auth.services); i++ {
		(*auth.services)[i].SetToken(token)
	}

	return c.String(http.StatusOK, "Logged in successfully!")
}

func (auth *AuthService) getAccessToken(state string, code string) (*oauth2.Token, error) {
	if state != auth.state {
		return nil, fmt.Errorf("invalid oauth state")
	}

	token, err := auth.microsoftAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}

	return token, nil
}

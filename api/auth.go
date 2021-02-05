package handler

import (
	"context"
	"log"
	"net/http"
	"sync"

	"golang.org/x/oauth2"

	"github.com/jbonadiman/finances-api/app_msgs"
	redisDB "github.com/jbonadiman/finances-api/databases/redis"
	"github.com/jbonadiman/finances-api/environment"
)

const (
	msAuthURL  = "https://login.microsoftonline.com/consumers/oauth2/v2.0/authorize"
	msTokenURL = "https://login.microsoftonline.com/consumers/oauth2/v2.0/token"
	tasksScope = "offline_access tasks.readwrite"
)

var (
	msConfig *oauth2.Config
)

func init() {
	consumerEndpoint := oauth2.Endpoint{}

	consumerEndpoint.AuthStyle = oauth2.AuthStyleInHeader
	consumerEndpoint.AuthURL = msAuthURL
	consumerEndpoint.TokenURL = msTokenURL

	msConfig = &oauth2.Config{
		RedirectURL:  environment.MSRedirectURL,
		ClientID:     environment.MSClientID,
		ClientSecret: environment.MSClientSecret,
		Scopes:       []string{ tasksScope },
		Endpoint:     consumerEndpoint,
	}
	// go workers.RequestAuthPage()
}

func StoreToken(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	authorizationCode := query.Get("code")

	log.Println("parsing authorize code from url query...")
	if authorizationCode == "" {
		app_msgs.SendBadRequest(&w, app_msgs.AuthCodeMissing())
		return
	}

	ctx := context.Background()

	log.Println("connecting to redis...")
	redisClient, err := redisDB.GetDB().GetClient()
	if err != nil {
		app_msgs.SendInternalError(&w, app_msgs.RedisConnectionError(err.Error()))
	}

	log.Println("retrieving token using authorize code...")
	token, err := msConfig.Exchange(ctx, authorizationCode)
	if err != nil {
		app_msgs.SendInternalError(&w, app_msgs.ErrorAuthenticating(err.Error()))
	}

	wg := sync.WaitGroup{}

	log.Println("storing token in cache...")

	wg.Add(3)
	go func() {
		redisClient.Set(
			ctx,
			"token:AccessToken",
			token.AccessToken,
			0,
		)

		wg.Done()
	}()

	go func() {
		redisClient.Set(
			ctx,
			"token:RefreshToken",
			token.RefreshToken,
			0,
		)

		wg.Done()
	}()

	go func() {
		redisClient.Set(
			ctx,
			"token:TokenType",
			token.TokenType,
			0,
		)

		wg.Done()
	}()

	// go func() {
	// 	redisClient.Set(
	// 		ctx,
	// 		"token:Expiry",
	// 		token.Expiry,
	// 		0,
	// 	)
	//
	// 	wg.Done()
	// }()

	wg.Wait()

	w.Write([]byte("Token stored successfully!"))
	return
}
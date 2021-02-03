package handler

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/jbonadiman/finance-bot/app_msgs"
	redisDB "github.com/jbonadiman/finance-bot/databases/redis"
	"github.com/jbonadiman/finance-bot/workers"
)

func StoreToken(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	authorizeCode := query.Get("code")
	state := query.Get("state")

	log.Println("validating authentication state...")
	if state == "" || state != workers.AuthState.String() {
		app_msgs.SendBadRequest(&w, app_msgs.InvalidAuthState(state))
	}

	log.Println("parsing authorize code from url query...")
	if authorizeCode == "" {
		app_msgs.SendBadRequest(&w, app_msgs.AuthCodeMissing())
		return
	}

	ctx := context.Background()

	log.Println("retrieving token using authorize code...")
	token, err := workers.MSConfig.Exchange(ctx, authorizeCode)
	if err != nil {
		app_msgs.SendInternalError(&w, app_msgs.ErrorAuthenticating(err.Error()))
	}

	redisClient, err := redisDB.GetDB().GetClient()
	if err != nil {
		app_msgs.SendInternalError(&w, app_msgs.RedisConnectionError(err.Error()))
	}

	log.Println("storing token in cache...")
	redisClient.Set(
		context.Background(),
		"token",
		token.AccessToken,
		token.Expiry.Sub(time.Now()),
	)
}
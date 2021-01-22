package handler

import (
	"context"
	"fmt"
	"golang.org/x/oauth2"
	"net/http"
)

func LoginRedirect(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	token, err := getAccessToken(query.Get("state"), query.Get("code"))
	if err != nil {
		fmt.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusInternalServerError)
	}

	_, _ = fmt.Fprintf(w, token.AccessToken)
}

func getAccessToken(state string, code string) (*oauth2.Token, error) {
	found := false

	for _, savedState := range statesList {
		if state == savedState {
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("invalid oauth states, try logging in again")
	}

	token, err := msConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}

	return token, nil
}

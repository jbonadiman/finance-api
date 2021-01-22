package handler

import (
	"context"
	"fmt"
	"golang.org/x/oauth2"
	"net/http"
)

func LoginRedirect(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	token, err := getAccessToken(query.Get("code"))
	if err != nil {
		fmt.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusInternalServerError)
	}

	_, _ = fmt.Fprintf(w, token.AccessToken)
}

func getAccessToken(code string) (*oauth2.Token, error) {
	token, err := msConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}

	return token, nil
}

package utils

import (
	"fmt"
	"log"
	"net/http"
)

func LoadVarSendingResponse(w *http.ResponseWriter, key string) string {
	msg := fmt.Sprintf("%q environment variable must be set!", key)

	envVar, err := LoadVar(key)
	if err != nil {
		log.Println(msg)
		http.Error(*w, msg, http.StatusBadRequest)
	}

	return envVar
}

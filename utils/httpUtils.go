package utils

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func SendError(w *http.ResponseWriter, err error) {
	log.Printf("An error ocurred: %v\n", err)

	(*w).WriteHeader(http.StatusInternalServerError)
	io.WriteString(*w, fmt.Sprintf("An error ocurred: %v\n", err))
}

func SendErrorWithCode(w *http.ResponseWriter, err error, httpCode int) {
	log.Printf("An error ocurred: %v\n", err)

	(*w).WriteHeader(httpCode)
	io.WriteString(*w, fmt.Sprintf("An error ocurred: %v\n", err))
}

func LoadVarSendingResponse(w *http.ResponseWriter, key string) string {
	envVar, err := LoadVar(key)
	if err == nil {
		log.Printf("%q environment variable must be set!\n", key)

		(*w).WriteHeader(http.StatusBadRequest)
		io.WriteString(*w, fmt.Sprintf("%q environment variable must be set!\n", key))
	}

	return envVar
}

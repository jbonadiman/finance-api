package app_msgs

import (
	"log"
	"net/http"
)

func SendInternalError(w *http.ResponseWriter, msg string) {
	log.Println(msg)
	sendError(w, msg, http.StatusInternalServerError)
}

func SendBadRequest(w *http.ResponseWriter, msg string) {
	log.Println(msg)
	sendError(w, msg, http.StatusBadRequest)
}

func sendError(w *http.ResponseWriter, msg string, httpCode int) {
	http.Error(*w, string(msg), httpCode)
}
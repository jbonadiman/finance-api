package main

import (
	"net/http"

	"github.com/jbonadiman/finance-bot/api"
)

func main() {
	http.HandleFunc("/api", handler.Index)
	http.HandleFunc("/api/get-tasks", handler.FetchTasks)

	http.ListenAndServe(":8080", nil)
}

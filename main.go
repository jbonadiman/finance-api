package main

import (
	handler "github.com/jbonadiman/finance-bot/api"
	"net/http"
)

func main() {
	http.HandleFunc("/api", handler.Index)
	http.HandleFunc("/api/get-tasks", handler.FetchTasks)

	http.ListenAndServe(":8080", nil)
}

package main

import (
	"net/http"

	"github.com/jbonadiman/finance-bot/api"
)

func main() {
	http.HandleFunc("/api/auth", handler.StoreToken)
	http.HandleFunc("/api/get-tasks", handler.FetchTasks)
	http.HandleFunc("/api/query", handler.QueryTransactions)

	http.ListenAndServe(":8080", nil)
}

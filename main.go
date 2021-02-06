package main

import (
	"net/http"

	"github.com/jbonadiman/finances-api/api"
)

func main() {
	http.HandleFunc(
		"/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("server is up"))
		},
	)

	http.HandleFunc("/api/auth", handler.StoreToken)
	http.HandleFunc("/api/get-tasks", handler.FetchTasks)
	http.HandleFunc("/api/query", handler.QueryTransactions)

	http.ListenAndServe(":8080", nil)
}

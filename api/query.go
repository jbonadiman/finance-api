package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/jbonadiman/finances-api/app_msgs"
	"github.com/jbonadiman/finances-api/databases/mongodb"
)

func init() {
	var err error

	mongoClient, err = mongodb.GetDB()
	if err != nil {
		log.Println(err.Error())
	}
}

func QueryTransactions(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	key := r.Header.Get("api_key")

	if !redisClient.CompareKeys(key) {
		log.Printf("non-authenticated call with key: %v\n", key)
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("Unauthorized request"))
		return
	}

	if subQuery := queryParams.Get("subcategory"); subQuery != "" {
		transactions, err := mongoClient.GetTransactionBySubcategory(subQuery)
		if err != nil {
			app_msgs.SendInternalError(&w, err.Error())
			return
		}

		transactionsAsJson, err := json.Marshal(transactions)
		if err != nil {
			app_msgs.SendInternalError(&w, err.Error())
			return
		}

		_, _ = w.Write(transactionsAsJson)
	}
}

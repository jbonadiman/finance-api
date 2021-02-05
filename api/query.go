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

		w.Write(transactionsAsJson)
	}
}

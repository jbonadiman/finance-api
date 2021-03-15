package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/jbonadiman/finances-api/internal/app_msgs"
	"github.com/jbonadiman/finances-api/internal/databases/mongodb"
	"github.com/jbonadiman/finances-api/internal/entities"
)

func init() {
	var err error

	mongoClient, err = mongodb.GetDB()
	if err != nil {
		log.Println(err.Error())
	}
}

func QueryTransactions(w http.ResponseWriter, r *http.Request) {
	user, password, ok := r.BasicAuth()

	if !ok || !redisClient.CompareAuthentication(user, password) {
		log.Printf("non-authenticated call with user:password: %q\n",
			user+":"+password)
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("Unauthorized request"))
		return
	}

	queryParams := r.URL.Query()

	var transactions *[]entities.Transaction
	var err error

	if allDataQuery := len(queryParams) == 0; allDataQuery {
		transactions, err = mongoClient.GetAllTransactions()
	} else if subcategoryQuery := queryParams.Get("subcategory"); subcategoryQuery != "" {
		transactions, err = mongoClient.GetTransactionBySubcategory(subcategoryQuery)
	}

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

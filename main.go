package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
)

const (
	clientId     = "MS_CLIENT_ID"
	clientSecret = "MS_CLIENT_SECRET"
	scope        = "MS_SCOPE"
	grantType    = "MS_GRANT_TYPE"
)

const (
	TaskListId = "AQMkADAwATNiZmYAZC1iNWMwLTQ3NDItMDACLTAwCgAuAAADY6fIEozObEqcJCMBbD9tYAEAPQLxMAsaBkSZbTEhjyRN5QAD5tJRHwAAAA=="
)

func main() {
	secret := "**REMOVED**"
	id := "**REMOVED**"

	credential, err := confidential.NewCredFromSecret(secret)
	if err != nil {
		log.Fatal(err)
	}

	client, err := confidential.New(id, credential)
	if err != nil {
		log.Fatal(err)
	}

	scopes := []string{"https://graph.microsoft.com/Tasks.ReadWrite"}

	token, err := client.AcquireTokenByCredential(context.Background(), scopes)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("GET", "https://graph.microsoft.com/v1.0/me/todo/lists", nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", token.AccessToken))

	httpClient := http.Client{Timeout: 10 * time.Second}

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	str, err := json.MarshalIndent(resp.Body, "", "    ")

	log.Println(string(str))

	//
	// var taskList = services.GetTasks(TaskListId)
	//
	// pool, err := utils.GetPool(2)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// repository := utils.NewMongoRepository(pool)
	//
	// for _, task := range *taskList {
	// 	splittedValues := strings.Split(task.Title, ";")
	//
	// 	var convertedValue float64
	//
	// 	convertedValue, err = strconv.ParseFloat(splittedValues[0], 64)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	//
	// 	transaction := entities.Transaction{
	// 		Date:        task.CreatedAt,
	// 		CreatedAt:   time.Now(),
	// 		ModifiedAt:  time.Now(),
	// 		Description: splittedValues[1],
	// 		Value:       convertedValue,
	// 		Category:    entities.Category{},
	// 	}
	//
	// 	err = repository.Store(&transaction)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	//
	// 	fmt.Println("title", task.Title)
	// 	fmt.Println("createdAt", task.CreatedAt)
	// 	fmt.Println("status", task.Status)
	// 	fmt.Println("============================")
	//
	// 	return
}

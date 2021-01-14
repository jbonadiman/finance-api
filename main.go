package main

import (
	"fmt"
	"github.com/jbonadiman/personal-finance-bot/services"
	"log"
	"strconv"
	"strings"
	"time"
)

const (
	TaskListId = "AQMkADAwATNiZmYAZC1iNWMwLTQ3NDItMDACLTAwCgAuAAADY6fIEozObEqcJCMBbD9tYAEAPQLxMAsaBkSZbTEhjyRN5QAD5tJRHwAAAA=="
)

func main() {
	var taskList = services.GetTasks(TaskListId)

	pool, err := GetPool(2)
	if err != nil {
		log.Fatal(err)
	}

	repository := NewMongoRepository(pool)

	for _, task := range *taskList {
		splittedValues := strings.Split(task.Title, ";")

		var convertedValue float64

		convertedValue, err = strconv.ParseFloat(splittedValues[0], 64)
		if err != nil {
			log.Fatal(err)
		}

		transaction := Transaction{
			Date:        task.CreatedAt,
			CreatedAt:   time.Now(),
			ModifiedAt:  time.Now(),
			Description: splittedValues[1],
			Value:       convertedValue,
			Category:    Category{},
		}

		err = repository.Store(&transaction)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("title", task.Title)
		fmt.Println("createdAt", task.CreatedAt)
		fmt.Println("status", task.Status)
		fmt.Println("============================")
		
		return
	}

}

package main

import (
	"fmt"
	"github.com/jbonadiman/personal-finance-bot/services"
)

const (
	TaskListId = "AQMkADAwATNiZmYAZC1iNWMwLTQ3NDItMDACLTAwCgAuAAADY6fIEozObEqcJCMBbD9tYAEAPQLxMAsaBkSZbTEhjyRN5QAD5tJRHwAAAA=="
)

func main() {
	var taskList = services.GetTasks(TaskListId)

	for _, task := range *taskList {
		fmt.Println("title", task.Title)
		fmt.Println("createdAt", task.CreatedAt)
		fmt.Println("status", task.Status)
		fmt.Println("============================")
	}

}

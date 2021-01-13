package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	MsTokenName = "MS_GRAPH_TOKEN"
)

const (
	TodoBaseUrl       = "https://graph.microsoft.com/v1.0/me/todo/lists/"
	TodoTasksUrl      = TodoBaseUrl + "%s/tasks"
	TodoDeleteTaskUrl = TodoBaseUrl + "%s/tasks/%s"
)

type TaskList struct {
	Value []Task `json:"value"`
}

type Task struct {
	Id           string    `json:"id"`
	Title        string    `json:"title"`
	CreatedAt    time.Time `json:"createdDateTime"`
	ModifiedAt   time.Time `json:"lastModifiedDateTime"`
	Importance   string    `json:"importance"`
	IsReminderOn bool      `json:"isReminderOn"`
	Status       string    `json:"status"`
}

func GetTasks(taskListId string) *[]Task {
	bearerToken := "Bearer " + os.Getenv(MsTokenName)

	tasksUrl := fmt.Sprintf(TodoTasksUrl, taskListId)

	req, err := http.NewRequest("GET", tasksUrl, nil)
	if err != nil {
		log.Fatal(err.Error())
	}

	req.Header.Add("Authorization", bearerToken)

	httpClient := &http.Client{Timeout: 10 * time.Second}

	resp, err := httpClient.Do(req)
	if err != nil || resp.StatusCode > http.StatusMultipleChoices {
		// TODO: improve this error message
		log.Fatal("The was an error sending the request")
	}

	var taskList TaskList

	err = json.NewDecoder(resp.Body).Decode(&taskList)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = resp.Body.Close()
	if err != nil {
		log.Fatal(err.Error())
	}

	for i := range taskList.Value {
		fixTimeZone(&taskList.Value[i])
	}

	return &taskList.Value
}

func fixTimeZone(task *Task) {
	saoPauloLocation, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		log.Fatal(err.Error())
	}

	task.CreatedAt = task.CreatedAt.In(saoPauloLocation)
	task.ModifiedAt = task.ModifiedAt.In(saoPauloLocation)
}

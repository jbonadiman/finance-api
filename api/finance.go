package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	TodoBaseUrl       = "https://graph.microsoft.com/v1.0/me/todo/lists/"
	TodoTasksUrl      = TodoBaseUrl + "%s/tasks"
	TodoDeleteTaskUrl = TodoBaseUrl + "%s/tasks/%s"
)

const (
	taskListEnv = "TASK_LIST_ID"
)

var (
	taskListId string
	client *http.Client
)

func init() {
	taskListId = os.Getenv(taskListEnv)

	if taskListId == "" {
		panic("Task List ID must be supplied!")
	}

	client = http.DefaultClient
}

type taskList struct {
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

func GetTasks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	token := query.Get("token")

	if token == "" {
		fmt.Fprintf(w, "Bearer token must be supplied!")
	}

	tasksUrl := fmt.Sprintf(TodoTasksUrl, taskListId)

	req, err := http.NewRequest("GET", tasksUrl, nil)
	if err != nil {
		fmt.Fprintf(w, "An error occurred when creating the request: %q", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", token))

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode > http.StatusMultipleChoices {
		fmt.Fprintf(w,"An error occurred sending the request: %v", err)
	}

	var tasks taskList

	err = json.NewDecoder(resp.Body).Decode(&tasks)
	if err != nil {
		fmt.Fprintf(w,
			"An error occurred deserializing the JSON: %v",
			err,
		)
	}

	err = resp.Body.Close()
	if err != nil {
		fmt.Fprintf(w,
			"An error occurred closing the request object: %v", err)
	}

	for i := range tasks.Value {
		fixTimeZone(&tasks.Value[i])
	}

	content, err := json.Marshal(tasks.Value)
	if err != nil {
		fmt.Fprintf(w,
			"An error occurred marshaling the json: %v", err)
	}

	fmt.Fprintf(w, string(content))
}

func fixTimeZone(task *Task) {
	locationName := "America/Sao_Paulo"

	saoPauloLocation, err := time.LoadLocation(locationName)
	if err != nil {
		log.Printf(
			"An error occurred loading the location %q: %v",
			locationName,
			err,
		)
	}

	task.CreatedAt = task.CreatedAt.In(saoPauloLocation)
	task.ModifiedAt = task.ModifiedAt.In(saoPauloLocation)
}

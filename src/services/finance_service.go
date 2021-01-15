package services

import (
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	MsTokenVarName = "MS_GRAPH_TOKEN"
)

const (
	TodoBaseUrl       = "https://graph.microsoft.com/v1.0/me/todo/lists/"
	TodoTasksUrl      = TodoBaseUrl + "%s/tasks"
	TodoDeleteTaskUrl = TodoBaseUrl + "%s/tasks/%s"
)

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

type Finance interface {
	Authorize(
		clientId string,
		clientSecret string,
		authUrl string,
		scope string) (*oauth2.Token, error)

	GetTasks(taskListId string) (*[]Task, error)
}

type FinanceService struct {}

func (f *FinanceService) Authorize(
	clientId string,
	clientSecret string,
	authUrl string,
	scope string) (*oauth2.Token, error) {

	return nil, nil
}

func (f *FinanceService) GetTasks(taskListId string) (*[]Task, error) {
	bearerToken := "Bearer " + os.Getenv(MsTokenVarName)

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

	var taskList taskList

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

	return &taskList.Value, nil
}

func fixTimeZone(task *Task) {
	saoPauloLocation, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		log.Fatal(err.Error())
	}

	task.CreatedAt = task.CreatedAt.In(saoPauloLocation)
	task.ModifiedAt = task.ModifiedAt.In(saoPauloLocation)
}

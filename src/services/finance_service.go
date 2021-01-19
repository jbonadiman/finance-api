package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/oauth2/clientcredentials"
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
	GetAuthorizedClient(
		clientId string,
		clientSecret string,
		authUrl string,
		scope string,
	) *http.Client

	GetTasks(taskListId string) (*[]Task, error)
}

type FinanceService struct {
	Logger     *log.Logger
	AuthClient *http.Client
}

func NewFinanceService(logger *log.Logger) *FinanceService {
	service := FinanceService{
		Logger: logger,
	}

	return &service
}

func (service *FinanceService) GetAuthorizedClient(
	clientId string,
	clientSecret string,
	authUrl string,
	scope string,
) *http.Client {
	oauthConfig := clientcredentials.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		TokenURL:     authUrl,
		Scopes:       []string{scope},
	}

	client := oauthConfig.Client(context.Background())
	client.Timeout = 10 * time.Second

	service.AuthClient = client

	return client
}

func (service *FinanceService) GetTasks(taskListId string) (*[]Task, error) {
	tasksUrl := fmt.Sprintf(TodoTasksUrl, taskListId)

	resp, err := service.AuthClient.Get(tasksUrl)
	if err != nil || resp.StatusCode > http.StatusMultipleChoices {
		service.Logger.Printf("An error occurred sending the request: %v", err)
		return nil, err
	}

	var tasks taskList

	err = json.NewDecoder(resp.Body).Decode(&tasks)
	if err != nil {
		service.Logger.Printf(
			"An error occurred deserializing the JSON: %v",
			err,
		)
		return nil, err
	}

	err = resp.Body.Close()
	if err != nil {
		service.Logger.Printf(
			"An error occurred closing the request object: %v",
			err,
		)
		return nil, err
	}

	for i := range tasks.Value {
		fixTimeZone(&tasks.Value[i])
	}

	return &tasks.Value, nil
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

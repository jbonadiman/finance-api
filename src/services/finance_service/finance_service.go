package finance_service

import (
	"encoding/json"
	"fmt"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/errors"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"time"
)

const (
	TodoBaseUrl       = "https://graph.microsoft.com/v1.0/me/todo/lists/"
	TodoTasksUrl      = TodoBaseUrl + "%s/tasks"
	TodoDeleteTaskUrl = TodoBaseUrl + "%s/tasks/%s"
)

const (
	TaskListId = "AQMkADAwATNiZmYAZC1iNWMwLTQ3NDItMDACLTAwCgAuAAADY6fIEozObEqcJCMBbD9tYAEAPQLxMAsaBkSZbTEhjyRN5QAD5tJRHwAAAA=="
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
	client *http.Client
	Token  *oauth2.Token
}

func New(client *http.Client) *FinanceService {
	service := FinanceService{
		client: client,
	}
	return &service
}

func (s *FinanceService) SetToken(token *oauth2.Token) {
	s.Token = token
}

func (s *FinanceService) GetTasks(c echo.Context) error {
	tasksUrl := fmt.Sprintf(TodoTasksUrl, TaskListId)

	if s.Token == nil {
		return errors.New("Not authenticated!")
	}

	req, err := http.NewRequest("GET", tasksUrl, nil)
	if err != nil {
		c.Logger().Printf("An error occurred when creating the request: %q", err)
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", s.Token.AccessToken))

	resp, err := s.client.Do(req)
	if err != nil || resp.StatusCode > http.StatusMultipleChoices {
		c.Logger().Printf("An error occurred sending the request: %v", err)
		return err
	}

	var tasks taskList

	err = json.NewDecoder(resp.Body).Decode(&tasks)
	if err != nil {
		c.Logger().Printf(
			"An error occurred deserializing the JSON: %v",
			err,
		)
		return err
	}

	err = resp.Body.Close()
	if err != nil {
		c.Logger().Printf(
			"An error occurred closing the request object: %v", err)
		return err
	}

	for i := range tasks.Value {
		fixTimeZone(&tasks.Value[i])
	}

	return c.JSON(http.StatusOK, &tasks.Value)
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

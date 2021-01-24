package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jbonadiman/finance-bot/databases/redis"
	"github.com/jbonadiman/finance-bot/models"
	"github.com/jbonadiman/finance-bot/utils"
	"log"
	"net/http"
	"time"
)

const (
	TodoBaseUrl       = "https://graph.microsoft.com/v1.0/me/todo/lists/"
	TodoTasksUrl      = TodoBaseUrl + "%v/tasks"
	TodoDeleteTaskUrl = TodoBaseUrl + "%v/tasks/%v"
)

var (
	TaskListID string
)

type taskList struct {
	Value []models.Task `json:"value"`
}

func init() {
	var err error

	TaskListID, err = utils.LoadVar("TASK_LIST_ID")
	if err != nil {
		log.Println(err.Error())
	}
}

func FetchTasks(w http.ResponseWriter, r *http.Request) {
	token, err := redis.GetTokenFromCache()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if token == "" {
		if MSClientID == "" || MSClientSecret == "" || MSRedirectUrl == "" {
			http.Error(w, "microsoft credentials environment variables must be set", http.StatusBadRequest)
		}

		queryCode := r.URL.Query().Get("code")
		if queryCode == "" {
			http.Error(w, "authorization code was not provided", http.StatusInternalServerError)
		}

		token, err = getCredentials(queryCode)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	tasks, err := getTasks(token)
	content, err := json.Marshal(tasks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(content)
}

func getCredentials(authorizationCode string) (string, error) {
	ctx := context.Background()

	token, err := MSConfig.Exchange(ctx, authorizationCode)
	if err != nil {
		return "", err
	}

	db, err := redis.New()
	if err != nil {
		return "", err
	}

	redisClient, err := db.GetClient()
	if err != nil {
		return "", err
	}

	redisClient.Set(
		context.Background(),
		"token",
		token.AccessToken,
		token.Expiry.Sub(time.Now()))

	return token.AccessToken, nil
}

func getTasks(token string) (*[]models.Task, error) {
	tasksUrl := fmt.Sprintf(TodoTasksUrl, TaskListID)

	req, err := http.NewRequest("GET", tasksUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	var tasks taskList

	err = json.NewDecoder(resp.Body).Decode(&tasks)
	if err != nil {
		return nil, err
	}

	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	for i := range tasks.Value {
		fixTimeZone(&tasks.Value[i])
	}

	return &tasks.Value, nil
}

func fixTimeZone(task *models.Task) {
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

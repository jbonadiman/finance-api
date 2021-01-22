package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jbonadiman/finance-bot/databases/redis"
	"github.com/jbonadiman/finance-bot/utils"
	"log"
	"net/http"
	"time"
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

func GetTasks(w http.ResponseWriter, r *http.Request) {
	token, err := getToken()
	if err != nil  {
		if err.Error() == "user not authenticated" {
			log.Println("User not authenticated, redirecting to the login page")
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tasks, err := getTasks(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	content, err := json.Marshal(tasks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(content)
}

func getTasks(token string) (*[]Task, error) {
	taskListId, err := utils.LoadVar("TASK_LIST_ID")
	if err != nil {
		return nil, err
	}

	tasksUrl := fmt.Sprintf(TodoTasksUrl, taskListId)

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

func getToken() (string, error) {
	ctx := context.Background()

	redisClient, err := redis.New().GetClient()
	if err != nil {
		log.Printf("Error creating Redis client: %v\n", err)
		return "", err
	}

	token := redisClient.Get(ctx, "token").Val()
	if token == "" {
		log.Println("User not authenticated!")

		return "", errors.New("user not authenticated")
	}

	return token, nil
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

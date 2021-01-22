package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jbonadiman/finance-bot/databases"
	"github.com/jbonadiman/finance-bot/utils"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	TodoBaseUrl       = "https://graph.microsoft.com/v1.0/me/todo/lists/"
	TodoTasksUrl      = TodoBaseUrl + "%s/tasks"
	TodoDeleteTaskUrl = TodoBaseUrl + "%s/tasks/%s"
)

var (
	taskListId string
	httpClient *http.Client
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

func init() {
	httpClient = http.DefaultClient
	cacheClient = databases.GetClient()

	var err error
	taskListId, err = utils.LoadVar("TASK_LIST_ID")
	if err != nil {
		log.Fatal(err.Error())
	}
}

func GetTasks(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	token := cacheClient.Get(ctx, "token").Val()
	if token == "" {
		log.Println("User not authenticated, redirecting to the login page")

		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	tasksUrl := fmt.Sprintf(TodoTasksUrl, taskListId)

	req, err := http.NewRequest("GET", tasksUrl, nil)
	if err != nil {
		utils.SendError(&w, err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", token))

	resp, err := httpClient.Do(req)
	if err != nil || resp.StatusCode > http.StatusMultipleChoices {
		utils.SendError(&w, err)
	}

	var tasks taskList

	err = json.NewDecoder(resp.Body).Decode(&tasks)
	if err != nil {
		utils.SendError(&w, err)
	}

	err = resp.Body.Close()
	if err != nil {
		utils.SendError(&w, err)
	}

	for i := range tasks.Value {
		fixTimeZone(&tasks.Value[i])
	}

	content, err := json.Marshal(tasks.Value)
	if err != nil {
		utils.SendError(&w, err)
	}

	io.WriteString(w, string(content))
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

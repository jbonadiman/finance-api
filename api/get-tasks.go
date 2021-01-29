package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/oauth2"

	"github.com/jbonadiman/finance-bot/databases/mongodb"
	redisdb "github.com/jbonadiman/finance-bot/databases/redis"
	"github.com/jbonadiman/finance-bot/entities"
	"github.com/jbonadiman/finance-bot/models"
	"github.com/jbonadiman/finance-bot/utils"
)

const (
	TodoBaseUrl       = "https://graph.microsoft.com/v1.0/me/todo/lists/"
	TodoTasksUrl      = TodoBaseUrl + "%v/tasks?$top=20"
	TodoDeleteTaskUrl = TodoBaseUrl + "%v/tasks/%v"

	locationName = "America/Sao_Paulo"
)

var (
	TaskListID  string
	mongoClient *mongodb.DB
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

	mongoClient, err = mongodb.New()
	if err != nil {
		log.Println(err.Error())
	}
}

func FetchTasks(w http.ResponseWriter, r *http.Request) {
	token, err := redisdb.GetTokenFromCache()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if token == "" {
		log.Println("checking for microsoft credentials in environment variables...")
		if MSClientID == "" || MSClientSecret == "" || MSRedirectUrl == "" {
			log.Println("microsoft credentials not found!")
			http.Error(
				w,
				"microsoft credentials environment variables must be set",
				http.StatusBadRequest,
			)
			return
		}

		log.Println("getting authorize code from url query...")
		queryCode := r.URL.Query().Get("code")
		if queryCode == "" {
			log.Println("could not find authorize code in url...")
			http.Error(
				w,
				"authorization code was not provided",
				http.StatusInternalServerError,
			)
			return
		}

		token, err = getCredentials(queryCode)
		if err != nil {
			log.Println("an error occurred while getting credentials...")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	tasks, err := getTasks(token)
	if err != nil {
		log.Printf(
			"an error occurred while retrieving tasks: %v\n",
			err.Error(),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	transactions, err := parseTasks(tasks)
	if err != nil {
		log.Printf("an error occurred while parsing tasks: %v\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	count, err := storeTransaction(transactions)
	if err != nil {
		log.Printf(
			"an error occurred while storing transactions: %v\n",
			err.Error(),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// err = deleteTasks(token, tasks)
	// if err != nil {
	// 	log.Printf(
	// 		"an error occurred deleting tasks: %v\n",
	// 		err.Error(),
	// 	)
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	w.Write([]byte(fmt.Sprintf("stored %v transactions successfully!", count)))
}

func getCredentials(authorizationCode string) (string, error) {
	ctx := context.Background()

	var token *oauth2.Token
	var redisClient *redis.Client
	var err error

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		log.Println("retrieving token using authorize code...")
		token, err = MSConfig.Exchange(ctx, authorizationCode)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		db, _ := redisdb.New()
		redisClient, err = db.GetClient()
		wg.Done()
	}()

	wg.Wait()

	log.Println("storing token in cache...")
	redisClient.Set(
		context.Background(),
		"token",
		token.AccessToken,
		token.Expiry.Sub(time.Now()),
	)

	return token.AccessToken, nil
}

func getTasks(token string) (*[]models.Task, error) {
	tasksUrl := fmt.Sprintf(TodoTasksUrl, TaskListID)

	req, err := http.NewRequest("GET", tasksUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", token))

	log.Println("listing tasks...")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var tasks taskList

	err = json.NewDecoder(resp.Body).Decode(&tasks)
	if err != nil {
		return nil, err
	}

	log.Printf("found %v tasks!", len(tasks.Value))

	return &tasks.Value, nil
}

func parseTasks(tasks *[]models.Task) (*[]entities.Transaction, error) {
	var transactions []entities.Transaction

	for _, task := range *tasks {
		values := strings.Split(task.Title, ";")

		cost, err := strconv.ParseFloat(
			strings.TrimSpace(values[0]),
			64,
		)
		if err != nil {
			return nil, err
		}

		description := strings.TrimSpace(values[1])
		category := strings.TrimSpace(values[2])

		transactions = append(
			transactions, entities.Transaction{
				ID:             primitive.NewObjectID(),
				Date:           task.CreatedAt,
				CreatedAt:      task.CreatedAt,
				ModifiedAt:     task.ModifiedAt,
				OriginalTaskID: task.Id,
				Description:    description,
				Cost:           cost,
				Category:       category,
			},
		)
	}

	return &transactions, nil
}

func storeTransaction(transactions *[]entities.Transaction) (int, error) {
	count, err := mongoClient.StoreTransactions(*transactions...)
	if err != nil {
		log.Printf(
			"an error ocurred. Stored %v transactions of %v: %v\n",
			count,
			len(*transactions),
			err.Error(),
		)
		return count, err
	} else {
		log.Printf(
			"all %v transactions were stored successfully!\n",
			count,
		)
		return count, nil
	}
}

func deleteTasks(token string, tasks *[]models.Task) error {
	var deleteUrls []*url.URL

	for _, task := range *tasks {
		urlDeleteTask, err := url.Parse(
			fmt.Sprintf(
				TodoDeleteTaskUrl,
				TaskListID,
				task.Id,
			),
		)

		if err != nil {
			return err
		}

		deleteUrls = append(
			deleteUrls,
			urlDeleteTask,
		)
	}

	wg := sync.WaitGroup{}

	authReq, err := http.NewRequest("DELETE", "", nil)
	if err != nil {
		return err
	}

	authReq.Header.Add("Authorization", fmt.Sprintf("Bearer %v", token))

	for _, u := range deleteUrls {
		wg.Add(1)
		go func(deleteUrl *url.URL) {
			log.Printf("executing request to %q\n", deleteUrl)

			newReq := authReq
			newReq.URL = deleteUrl

			_, _ = http.DefaultClient.Do(newReq)
			wg.Done()
		}(u)
	}

	wg.Wait()

	log.Printf("deleted %v tasks!", len(*tasks))
	return nil
}

package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/jbonadiman/finance-bot/app_msgs"
	"github.com/jbonadiman/finance-bot/databases/mongodb"
	redisDB "github.com/jbonadiman/finance-bot/databases/redis"
	"github.com/jbonadiman/finance-bot/entities"
	"github.com/jbonadiman/finance-bot/environment"
	"github.com/jbonadiman/finance-bot/models"
)

const (
	TodoBaseUrl       = "https://graph.microsoft.com/v1.0/me/todo/lists/"
	TodoTasksUrl      = TodoBaseUrl + "%v/tasks?$top=20"
	TodoDeleteTaskUrl = TodoBaseUrl + "%v/tasks/%v"
)

var (
	mongoClient *mongodb.DB
	httpClient *http.Client
)

type taskList struct {
	Value []models.Task `json:"value"`
}

func init() {
	var err error

	mongoClient, err = mongodb.GetDB()
	if err != nil {
		log.Println(err.Error())
	}

	token, err := redisDB.GetTokenFromCache()
	if err != nil {
		log.Fatalf("could not retrieve token: %v\n", err.Error())
	}

	if token != nil {
		httpClient = msConfig.Client(context.Background(), token)
	} else {
		log.Println("token is not on cache yet")
	}
}

func FetchTasks(w http.ResponseWriter, _ *http.Request) {
	if httpClient == nil {
		token, err := redisDB.GetTokenFromCache()
		if err != nil {
			log.Printf("could not retrieve token: %v\n", err.Error())
			app_msgs.SendInternalError(&w, err.Error())
		}

		if token != nil {
			httpClient = msConfig.Client(context.Background(), token)
		} else {
			log.Println("could not assemble token")
			app_msgs.SendBadRequest(&w, app_msgs.NotAuthenticated())
		}
	}

	tasks, err := getTasks()
	if err != nil {
		log.Println("an error occurred while retrieving tasks...")
		app_msgs.SendInternalError(&w, err.Error())
		return
	}

	if len(*tasks) == 0 {
		w.Write([]byte("could not find any tasks to be stored"))
		return
	}

	transactions, err := parseTasks(tasks)
	if err != nil {
		log.Println("an error occurred while parsing tasks...")
		app_msgs.SendBadRequest(&w, err.Error())
		return
	}

	count, err := storeTransaction(transactions)
	if err != nil {
		log.Println("an error occurred while storing transactions...")
		app_msgs.SendInternalError(&w, err.Error())
		return
	}

	err = deleteTasks(tasks)
	if err != nil {
		app_msgs.SendInternalError(&w, app_msgs.ErrorDeletingTasks(err.Error()))
		return
	}

	w.Write([]byte(fmt.Sprintf("stored %v transactions successfully!", count)))
}

func getTasks() (*[]models.Task, error) {
	var tasks taskList

	tasksUrl := fmt.Sprintf(TodoTasksUrl, environment.TaskListID)

	req, err := http.NewRequest("GET", tasksUrl, nil)
	if err != nil {
		return nil, err
	}

	log.Println("listing tasks...")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&tasks)
	if err != nil {
		return nil, err
	}

	log.Printf("found %v tasks!", len(tasks.Value))

	return &tasks.Value, nil
}

func parseTasks(tasks *[]models.Task) (*[]entities.Transaction, error) {
	transactions := make([]entities.Transaction, len(*tasks))

	wg := sync.WaitGroup{}

	for i, task := range *tasks {
		wg.Add(1)
		go func(index int, t models.Task) {
			defer wg.Done()
			values := strings.Split(t.Title, ";")

			cost, err := strconv.ParseFloat(
				strings.TrimSpace(values[0]),
				64,
			)

			if err != nil {
				return
			}

			description := strings.TrimSpace(values[1])
			unparsedCategory := strings.TrimSpace(values[2])

			category, err := parseSubcategory(unparsedCategory)
			if err != nil {
				return
			}

			transactions[index] = entities.Transaction{
				ID:             primitive.NewObjectID(),
				Date:           t.CreatedAt,
				CreatedAt:      t.CreatedAt,
				ModifiedAt:     t.ModifiedAt,
				OriginalTaskID: t.Id,
				Description:    description,
				Cost:           cost,
				Subcategory:    category,
			}

		}(i, task)
	}

	wg.Wait()

	parsed := len(transactions)
	total := len(*tasks)

	if parsed != total {
		return &transactions, errors.New(app_msgs.NotAllTasksParsed(parsed, total))
	}

	return &transactions, nil
}

func parseSubcategory(sub string) (string, error) {
	subcategory, err := mongoClient.ParseCategory(sub)
	if err != nil {
		return "", err
	}

	return (*subcategory).Name, nil
}

func storeTransaction(transactions *[]entities.Transaction) (int, error) {
	count, err := mongoClient.StoreTransactions(*transactions...)
	if err != nil {
		log.Println(app_msgs.NotAllTransactionsStored(count, len(*transactions)))
		return count, err
	}

	log.Printf(app_msgs.AllTransactionsStored(count))
	return count, nil
}

func deleteTasks(tasks *[]models.Task) error {
	authReq, err := http.NewRequest("DELETE", "", nil)
	if err != nil {
		return err
	}

	for _, task := range *tasks {
		urlDeleteTask, err := url.Parse(
			fmt.Sprintf(
				TodoDeleteTaskUrl,
				environment.TaskListID,
				task.Id,
			),
		)

		if err != nil {
			return err
		}

		log.Printf("executing request to %q\n", urlDeleteTask)

		newReq := authReq
		newReq.URL = urlDeleteTask

		_, err = httpClient.Do(newReq)
		if err != nil {
			return err
		}
	}

	log.Printf(app_msgs.AllTasksDeleted(len(*tasks)))
	return nil
}

package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/oauth2"

	"github.com/jbonadiman/finances-api/internal/app_msgs"
	"github.com/jbonadiman/finances-api/internal/databases/mongodb"
	redisDB "github.com/jbonadiman/finances-api/internal/databases/redis"
	"github.com/jbonadiman/finances-api/internal/entities"
	"github.com/jbonadiman/finances-api/internal/environment"
	"github.com/jbonadiman/finances-api/internal/models"
)

const (
	BaseUrl       = "https://graph.microsoft.com/v1.0/me/todo/lists/"
	FetchTasksUrl = BaseUrl + "%v/tasks?$filter=status%%20eq%%20'notStarted'&$top=20"
	AlterTaskUrl  = BaseUrl + "%v/tasks/%v"
)

var (
	mongoClient *mongodb.DB
	redisClient *redisDB.DB

	httpClient *http.Client

	token       *oauth2.Token
	tokenSource oauth2.TokenSource
)

type taskList struct {
	Value []models.Task `json:"value"`
}

func init() {
	var err error

	log.Println("connecting to mongoDB...")
	mongoClient, err = mongodb.GetDB()
	if err != nil {
		log.Fatalf(err.Error())
	}

	log.Println("connecting to redis...")
	redisClient, err = redisDB.GetDB()
	if err != nil {
		log.Fatalf(err.Error())
	}

	token, err = redisClient.GetToken()
	if err != nil {
		log.Fatalf("could not retrieve token: %v\n", err.Error())
	}

	ctx := context.Background()

	tokenSource = msConfig.TokenSource(ctx, token)
	httpClient = msConfig.Client(ctx, token)
}

func FetchTasks(w http.ResponseWriter, r *http.Request) {
	user, password, ok := r.BasicAuth()

	if !ok || !redisClient.CompareAuthentication(user, password) {
		log.Printf(
			"non-authenticated call with user:password: %q\n",
			user+":"+password,
		)
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("Unauthorized request"))
		return
	}

	storeRefreshedToken()

	tasks, err := getNotStartedTasks()
	if err != nil {
		log.Println("an error occurred while retrieving tasks...")
		app_msgs.SendInternalError(&w, err.Error())
		return
	}

	if len(*tasks) == 0 {
		_, _ = w.Write([]byte("could not find any tasks to be stored"))
		return
	}

	transactions, errList := parseTasks(tasks)
	if len(errList) > 0 {
		log.Println("could not parse all tasks:")
		for _, e := range errList {
			log.Println(e.Error())
		}
	}

	count, err := storeTransaction(transactions)
	if err != nil {
		log.Println("an error occurred while storing transactions...")
		app_msgs.SendInternalError(&w, err.Error())
		return
	}

	err = markTasksAsCompleted(tasks)
	if err != nil {
		app_msgs.SendInternalError(
			&w,
			app_msgs.ErrorCompletingTasks(err.Error()),
		)
		return
	}

	_, _ = w.Write(
		[]byte(fmt.Sprintf(
			"stored %v transactions successfully!",
			count,
		)),
	)
}

func storeRefreshedToken() {
	newToken, err := tokenSource.Token()
	if err != nil {
		log.Fatalln(err)
	}

	if newToken.AccessToken != token.AccessToken {
		wg := sync.WaitGroup{}

		wg.Add(2)
		go func() {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), TimeOut)
			defer cancel()

			token = newToken
			tokenSource = msConfig.TokenSource(ctx, token)

			httpClient = msConfig.Client(ctx, newToken)
		}()

		go func() {
			defer wg.Done()
			redisClient.StoreToken(newToken)
		}()

		wg.Wait()
		log.Println("token refreshed successfully")
	}
}

func getNotStartedTasks() (*[]models.Task, error) {
	var tasks taskList

	tasksUrl := fmt.Sprintf(FetchTasksUrl, environment.TaskListID)

	log.Printf("listing tasks using url %q...\n", tasksUrl)
	resp, err := httpClient.Get(tasksUrl)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		log.Printf(
			"unsuccessful request (status code '%v'). retrieving body...\n",
			resp.StatusCode,
		)

		var bodyBytes []byte

		bodyBytes, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return nil, errors.New(string(bodyBytes))
	}

	log.Println("successful request. parsing body to JSON...")
	err = json.NewDecoder(resp.Body).Decode(&tasks)
	if err != nil {
		return nil, err
	}

	log.Printf("found %v tasks!\n", len(tasks.Value))

	return &tasks.Value, nil
}

func parseTasks(tasks *[]models.Task) (*[]entities.Transaction, []error) {
	transactions := make([]entities.Transaction, len(*tasks))
	errorList := make([]error, 0)

	wg := sync.WaitGroup{}

	for i, task := range *tasks {
		wg.Add(1)
		go func(index int, t models.Task) {
			defer wg.Done()
			transactionParts := strings.Split(t.Title, ";")

			if len(transactionParts) != 3 {
				errorList = append(
					errorList,
					errors.New(
						fmt.Sprintf(
							"the task: %q is invalid. A transaction must be composed of three parts: one for the cost, one for the description and another for the category",
							t.Title,
						),
					),
				)
				return
			}


			cost, err := strconv.ParseFloat(
				strings.TrimSpace(transactionParts[0]),
				64,
			)

			if err != nil || cost <= 0 {
				errorList = append(
					errorList,
					errors.New(
						fmt.Sprintf(
							"cost value in task %q is invalid",
							t.Title,
						),
					),
				)
				return
			}

			description := strings.TrimSpace(transactionParts[1])
			unparsedCategory := strings.TrimSpace(transactionParts[2])

			if description == "" {
				errorList = append(
					errorList,
					errors.New(
						fmt.Sprintf(
							"task %q has no description",
							t.Title,
						),
					),
				)
				return
			}

			subcategory, err := redisClient.ParseSubcategory(unparsedCategory)
			if err != nil || subcategory == "" {
				errorList = append(
					errorList,
					errors.New(
						fmt.Sprintf(
							"could not parse subcategory of task %q",
							t.Title,
						),
					),
				)
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
				Subcategory:    subcategory,
			}

		}(i, task)
	}

	wg.Wait()

	if len(errorList) > 0 {
		return &transactions, errorList
	}

	return &transactions, nil
}

func storeTransaction(transactions *[]entities.Transaction) (int, error) {
	count, err := mongoClient.StoreTransactions(*transactions...)
	if err != nil {
		log.Println(
			app_msgs.NotAllTransactionsStored(
				count,
				len(*transactions),
			),
		)
		return count, err
	}

	log.Printf(app_msgs.AllTransactionsStored(count))
	return count, nil
}

func markTasksAsCompleted(tasks *[]models.Task) error {
	for _, task := range *tasks {
		urlTask :=
			fmt.Sprintf(
				AlterTaskUrl,
				environment.TaskListID,
				task.Id,
			)

		log.Printf("executing request to %q\n", urlTask)

		newReq, err := http.NewRequest(
			http.MethodPatch,
			urlTask,
			strings.NewReader("{\"status\":\"completed\"}"),
		)
		if err != nil {
			return err
		}

		newReq.Header.Set("Content-Type", "application/json")
		resp, err := httpClient.Do(newReq)
		if err != nil {
			return err
		}

		if resp.StatusCode >= 400 {
			log.Printf(
				"unsuccessful request (status code '%v'). retrieving body...\n",
				resp.StatusCode,
			)

			var bodyBytes []byte

			bodyBytes, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			return errors.New(string(bodyBytes))
		}

		resp.Body.Close()
	}

	log.Printf(app_msgs.AllTasksCompleted(len(*tasks)))
	return nil
}

func deleteTasks(tasks *[]models.Task) error {
	authReq, err := http.NewRequest(http.MethodDelete, "", nil)
	if err != nil {
		return err
	}

	for _, task := range *tasks {
		urlDeleteTask, err := url.Parse(
			fmt.Sprintf(
				AlterTaskUrl,
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

	log.Printf(app_msgs.AllTasksCompleted(len(*tasks)))
	return nil
}

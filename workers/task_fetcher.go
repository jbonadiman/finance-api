package workers
//
// import (
// 	"context"
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"net/url"
// 	"strconv"
// 	"strings"
// 	"sync"
// 	"time"
//
// 	"github.com/go-redis/redis/v8"
// 	"go.mongodb.org/mongo-driver/bson/primitive"
// 	"golang.org/x/oauth2"
//
// 	"github.com/jbonadiman/finance-bot/app_msgs"
// 	"github.com/jbonadiman/finance-bot/databases/mongodb"
// 	redisDB "github.com/jbonadiman/finance-bot/databases/redis"
// 	"github.com/jbonadiman/finance-bot/entities"
// 	"github.com/jbonadiman/finance-bot/models"
// 	"github.com/jbonadiman/finance-bot/utils"
// )
//
// const (
// 	TodoBaseUrl       = "https://graph.microsoft.com/v1.0/me/todo/lists/"
// 	TodoTasksUrl      = TodoBaseUrl + "%v/tasks?$top=20"
// 	TodoDeleteTaskUrl = TodoBaseUrl + "%v/tasks/%v"
// )
//
// var (
// 	TaskListID  string
// 	mongoClient *mongodb.DB
// )
//
// type taskList struct {
// 	Value []models.Task `json:"value"`
// }
//
// func init() {
// 	var err error
//
// 	TaskListID, err = utils.LoadVar("TASK_LIST_ID")
// 	if err != nil {
// 		log.Println(err.Error())
// 	}
//
// 	mongoClient, err = mongodb.New()
// 	if err != nil {
// 		log.Println(err.Error())
// 	}
// }
//
// func FetchTasks(w http.ResponseWriter, r *http.Request) {
// 	token, err := redisDB.GetTokenFromCache()
// 	if err != nil {
// 		app_msgs.SendInternalError(&w, err.Error())
// 		return
// 	}
//
// 	if token == "" {
// 		log.Println("checking for microsoft credentials in environment variables...")
// 		if MSClientID == "" || MSClientSecret == "" || MSRedirectUrl == "" {
// 			app_msgs.SendBadRequest(&w, app_msgs.MsCredentials())
// 			return
// 		}
//
// 		log.Println("getting authorize code from url query...")
// 		queryCode := r.URL.Query().Get("code")
// 		if queryCode == "" {
// 			app_msgs.SendBadRequest(&w, app_msgs.AuthCodeMissing())
// 			return
// 		}
//
// 		token, err = getCredentials(queryCode)
// 		if err != nil {
// 			log.Println("an error occurred while getting credentials...")
// 			app_msgs.SendInternalError(&w, err.Error())
// 			return
// 		}
// 	}
//
// 	tasks, err := getTasks(token)
// 	if err != nil {
// 		log.Println("an error occurred while retrieving tasks...")
// 		app_msgs.SendInternalError(&w, err.Error())
// 		return
// 	}
//
// 	if len(*tasks) == 0 {
// 		w.Write([]byte("could not find any tasks to be stored"))
// 		return
// 	}
//
// 	transactions, err := parseTasks(tasks, mongoClient)
// 	if err != nil {
// 		log.Println("an error occurred while parsing tasks...")
// 		app_msgs.SendBadRequest(&w, err.Error())
// 		return
// 	}
//
// 	count, err := storeTransaction(transactions)
// 	if err != nil {
// 		log.Println("an error occurred while storing transactions...")
// 		app_msgs.SendInternalError(&w, err.Error())
// 		return
// 	}
//
// 	err = deleteTasks(token, tasks)
// 	if err != nil {
// 		app_msgs.SendInternalError(&w, app_msgs.ErrorDeletingTasks(err.Error()))
// 		return
// 	}
//
// 	w.Write([]byte(fmt.Sprintf("stored %v transactions successfully!", count)))
// }
//
// func getCredentials(authorizationCode string) (string, error) {
// 	ctx := context.Background()
//
// 	var token *oauth2.Token
// 	var redisClient *redis.Client
// 	var err error
//
// 	wg := sync.WaitGroup{}
//
// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		log.Println("retrieving token using authorize code...")
// 		token, err = MSConfig.Exchange(ctx, authorizationCode)
// 	}()
//
// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		db, _ := redisDB.New()
// 		redisClient, err = db.GetClient()
// 	}()
//
// 	wg.Wait()
//
// 	if err != nil {
// 		return "", err
// 	}
//
// 	log.Println("storing token in cache...")
// 	redisClient.Set(
// 		context.Background(),
// 		"token",
// 		token.AccessToken,
// 		token.Expiry.Sub(time.Now()),
// 	)
//
// 	return token.AccessToken, nil
// }
//
// func getTasks(token string) (*[]models.Task, error) {
// 	var tasks taskList
//
// 	tasksUrl := fmt.Sprintf(TodoTasksUrl, TaskListID)
//
// 	req, err := http.NewRequest("GET", tasksUrl, nil)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", token))
//
// 	log.Println("listing tasks...")
// 	resp, err := http.DefaultClient.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	defer resp.Body.Close()
//
// 	err = json.NewDecoder(resp.Body).Decode(&tasks)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	log.Printf("found %v tasks!", len(tasks.Value))
//
// 	return &tasks.Value, nil
// }
//
// func parseTasks(
// 	tasks *[]models.Task,
// 	mongo *mongodb.DB,
// ) (*[]entities.Transaction, error) {
// 	transactions := make([]entities.Transaction, len(*tasks))
//
// 	wg := sync.WaitGroup{}
//
// 	for i, task := range *tasks {
// 		wg.Add(1)
// 		go func(index int, t models.Task) {
// 			defer wg.Done()
// 			values := strings.Split(t.Title, ";")
//
// 			cost, err := strconv.ParseFloat(
// 				strings.TrimSpace(values[0]),
// 				64,
// 			)
//
// 			if err != nil {
// 				return
// 			}
//
// 			description := strings.TrimSpace(values[1])
// 			unparsedCategory := strings.TrimSpace(values[2])
//
// 			category, err := parseSubcategory(unparsedCategory, mongo)
// 			if err != nil {
// 				return
// 			}
//
// 			transactions[index] = entities.Transaction{
// 				ID:             primitive.NewObjectID(),
// 				Date:           t.CreatedAt,
// 				CreatedAt:      t.CreatedAt,
// 				ModifiedAt:     t.ModifiedAt,
// 				OriginalTaskID: t.Id,
// 				Description:    description,
// 				Cost:           cost,
// 				Subcategory:    category,
// 			}
//
// 		}(i, task)
// 	}
//
// 	wg.Wait()
//
// 	parsed := len(transactions)
// 	total := len(*tasks)
//
// 	if parsed != total {
// 		return &transactions, errors.New(app_msgs.NotAllTasksParsed(parsed, total))
// 	}
//
// 	return &transactions, nil
// }
//
// func parseSubcategory(sub string, mongo *mongodb.DB) (string, error) {
// 	subcategory, err := mongo.ParseCategory(sub)
// 	if err != nil {
// 		return "", err
// 	}
//
// 	return (*subcategory).Name, nil
// }
//
// func storeTransaction(transactions *[]entities.Transaction) (int, error) {
// 	count, err := mongoClient.StoreTransactions(*transactions...)
// 	if err != nil {
// 		log.Println(app_msgs.NotAllTransactionsStored(count, len(*transactions)))
// 		return count, err
// 	}
//
// 	log.Printf(app_msgs.AllTransactionsStored(count))
// 	return count, nil
// }
//
// func deleteTasks(token string, tasks *[]models.Task) error {
// 	authReq, err := http.NewRequest("DELETE", "", nil)
// 	if err != nil {
// 		return err
// 	}
//
// 	authReq.Header.Add("Authorization", fmt.Sprintf("Bearer %v", token))
//
// 	for _, task := range *tasks {
// 		urlDeleteTask, err := url.Parse(
// 			fmt.Sprintf(
// 				TodoDeleteTaskUrl,
// 				TaskListID,
// 				task.Id,
// 			),
// 		)
//
// 		if err != nil {
// 			return err
// 		}
//
// 		log.Printf("executing request to %q\n", urlDeleteTask)
//
// 		newReq := authReq
// 		newReq.URL = urlDeleteTask
//
// 		_, err = http.DefaultClient.Do(newReq)
// 		if err != nil {
// 			return err
// 		}
// 	}
//
// 	log.Printf(app_msgs.AllTasksDeleted(len(*tasks)))
// 	return nil
// }

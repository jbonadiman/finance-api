package main

import (
	"github.com/jbonadiman/finance-bot/src/services"
	"github.com/jbonadiman/finance-bot/src/services/auth_service"
	"github.com/jbonadiman/finance-bot/src/services/finance_service"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/oauth2"
	"net/http"
	"os"
	"time"
)

const (
	clientId     = "MS_CLIENT_ID"
	clientSecret = "MS_CLIENT_SECRET"
	scope        = "MS_SCOPE"
	grantType    = "MS_GRANT_TYPE"
)



var (
	oauthStateString          string
	microsoftOAuthConfig      *oauth2.Config
	microsoftConsumerEndpoint oauth2.Endpoint
)

func init() {
	microsoftConsumerEndpoint.AuthStyle = oauth2.AuthStyleInHeader
	microsoftConsumerEndpoint.AuthURL = "https://login.microsoftonline.com/consumers/oauth2/v2.0/authorize"
	microsoftConsumerEndpoint.TokenURL = "https://login.microsoftonline.com/consumers/oauth2/v2.0/token"

	microsoftOAuthConfig = &oauth2.Config{
		RedirectURL:  "http://localhost:8080/authentication",
		ClientID:     os.Getenv("MS_CLIENT_ID"),
		ClientSecret: os.Getenv("MS_CLIENT_SECRET"),
		Scopes:       []string{"offline_access tasks.readwrite"},
		Endpoint:     microsoftConsumerEndpoint,
	}
}

func main() {
	e := echo.New()

	client := http.Client{Timeout: 10 * time.Second}

	financeService := finance_service.New(&client)
	authService := auth_service.New(&[]services.Authenticated{financeService})

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", healthCheck)

	e.GET("/login", authService.Login)
	e.GET("/authentication", authService.LoginRedirect)
	e.GET("/tasks", financeService.GetTasks)

	e.Logger.Fatal(e.Start(":8080"))

	//req, err := http.NewRequest("GET", "https://graph.microsoft.com/v1.0/me/todo/lists", nil)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", token.AccessToken))
	//
	//httpClient := http.Client{Timeout: 10 * time.Second}
	//
	//resp, err := httpClient.Do(req)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//str, err := json.MarshalIndent(resp.Body, "", "    ")
	//
	//log.Println(string(str))

	//
	// var taskList = services.GetTasks(TaskListId)
	//
	// pool, err := utils.GetPool(2)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// repository := utils.NewMongoRepository(pool)
	//
	// for _, task := range *taskList {
	// 	splittedValues := strings.Split(task.Title, ";")
	//
	// 	var convertedValue float64
	//
	// 	convertedValue, err = strconv.ParseFloat(splittedValues[0], 64)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	//
	// 	transaction := entities.Transaction{
	// 		Date:        task.CreatedAt,
	// 		CreatedAt:   time.Now(),
	// 		ModifiedAt:  time.Now(),
	// 		Description: splittedValues[1],
	// 		Value:       convertedValue,
	// 		Category:    entities.Category{},
	// 	}
	//
	// 	err = repository.Store(&transaction)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	//
	// 	fmt.Println("title", task.Title)
	// 	fmt.Println("createdAt", task.CreatedAt)
	// 	fmt.Println("status", task.Status)
	// 	fmt.Println("============================")
	//
	// 	return
}

func healthCheck(c echo.Context) error {
	return c.String(http.StatusOK, "")
}

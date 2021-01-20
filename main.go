package main

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"net/http"
	"os"
)

const (
	clientId     = "MS_CLIENT_ID"
	clientSecret = "MS_CLIENT_SECRET"
	scope        = "MS_SCOPE"
	grantType    = "MS_GRANT_TYPE"
)

const (
	TaskListId = "AQMkADAwATNiZmYAZC1iNWMwLTQ3NDItMDACLTAwCgAuAAADY6fIEozObEqcJCMBbD9tYAEAPQLxMAsaBkSZbTEhjyRN5QAD5tJRHwAAAA=="
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
	//http.HandleFunc("/", handleMain)
	http.HandleFunc("/login", handleMicrosoftLogin)
	http.HandleFunc("/authentication", handleMicrosoftCallback)

	http.ListenAndServe(":8080", nil)


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

func handleMicrosoftLogin(w http.ResponseWriter, r *http.Request) {
	oauthStateString = uuid.New().String()

	url := microsoftOAuthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleMicrosoftCallback(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	token, err := getAccessToken(query.Get("state"), query.Get("code"))
	if err != nil {
		fmt.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	fmt.Fprintf(w, "Access token: %s\n", token)
}

func getAccessToken(state string, code string) (*oauth2.Token, error) {
	if state != oauthStateString {
		return nil, fmt.Errorf("invalid oauth state")
	}

	token, err := microsoftOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}

	return token, nil
	//response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	//if err != nil {
	//	return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	//}
	//defer response.Body.Close()
	//contents, err := ioutil.ReadAll(response.Body)
	//if err != nil {
	//	return nil, fmt.Errorf("failed reading response body: %s", err.Error())
	//}
	//return contents, nil
}

package environment

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
)

var doOnce sync.Once

var (
	MSClientID     string
	MSClientSecret string
	MSRedirectURL  string
)

var (
	LambdaHost string
	LambdaPassword  string
	LambdaPort      string
)

var (
	MongoHost     string
	MongoPassword string
	MongoUser     string
)

var (
	TaskListID string
)

func init() {
	loadMicrosoftVars()
	loadLambdaStoreVars()
	loadMongoDBVars()

	var err error
	TaskListID, err = loadVar("TASK_LIST_ID")
	if err != nil {
		log.Fatal(err.Error())
	}
}

func loadLambdaStoreVars() {
	var err error

	doOnce.Do(func() {
		log.Println("loading lambda store environment variables...")

		LambdaHost, err = loadVar("LAMBDA_HOST")
		if err != nil {
			log.Fatal(err.Error())
		}

		LambdaPassword, err = loadVar("LAMBDA_SECRET")
		if err != nil {
			log.Fatal(err.Error())
		}

		LambdaPort, err = loadVar("LAMBDA_PORT")
		if err != nil {
			log.Fatal(err.Error())
		}
	})
}

func loadMicrosoftVars() {
	var err error

	doOnce.Do(func() {
		log.Println("loading microsoft environment variables...")

		MSClientID, err = loadVar("MS_CLIENT_ID")
		if err != nil {
			log.Fatal(err.Error())
		}

		MSClientSecret, err = loadVar("MS_CLIENT_SECRET")
		if err != nil {
			log.Fatal(err.Error())
		}

		MSRedirectURL, err = loadVar("MS_REDIRECT")
		if err != nil {
			log.Fatal(err.Error())
		}
	})
}

func loadMongoDBVars() {
	var err error

	doOnce.Do(func() {
		log.Println("loading mongodb atlas environment variables...")

		MongoHost, err = loadVar("MONGO_HOST")
		if err != nil {
			log.Fatal(err.Error())
		}

		MongoPassword, err = loadVar("MONGO_SECRET")
		if err != nil {
			log.Fatal(err.Error())
		}

		MongoUser, err = loadVar("MONGO_USER")
		if err != nil {
			log.Fatal(err.Error())
		}
	})
}


func loadVar(key string) (string, error) {
	variable := os.Getenv(key)
	if variable == "" {
		return "", errors.New(fmt.Sprintf("%q environment variable not set!", key))
	}

	return variable, nil
}

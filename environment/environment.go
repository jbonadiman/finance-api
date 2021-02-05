package environment

import (
	"errors"
	"fmt"
	"log"
	"os"
)

const (
	LambdaHostKey            = "LAMBDA_HOST"
	LambdaSecretKey          = "LAMBDA_SECRET"
	LambdaPortKey            = "LAMBDA_PORT"
	MicrosoftClientIDKey     = "MS_CLIENT_ID"
	MicrosoftClientSecretKey = "MS_CLIENT_SECRET"
	MicrosoftRedirectURLKey  = "MS_REDIRECT"
	MongoHostKey             = "MONGO_HOST"
	MongoUserKey             = "MONGO_USER"
	MongoPasswordKey         = "MONGO_SECRET"

	TaskListIDKey = "TASK_LIST_ID"
	ReadOnlyTasksKey = "READONLY_TASKS"
)

var (
	MSClientID     string
	MSClientSecret string
	MSRedirectURL  string
)

var (
	LambdaHost     string
	LambdaPassword string
	LambdaPort     string
)

var (
	MongoHost     string
	MongoPassword string
	MongoUser     string
)

var (
	TaskListID    string
	ReadOnlyTasks string
)

func init() {
	unsetVarList := make([]string, 0)

	unsetVarList = append(unsetVarList, loadMicrosoftVars()...)
	unsetVarList = append(unsetVarList, loadLambdaStoreVars()...)
	unsetVarList = append(unsetVarList, loadMongoDBVars()...)

	if len(unsetVarList) > 0 {
		for _, err := range unsetVarList {
			log.Println(err)
		}

		os.Exit(1)
	}

	var err error

	TaskListID, err = loadVar(TaskListIDKey)
	if err != nil {
		log.Fatal(err.Error())
	}

	ReadOnlyTasks, _ = loadVar(ReadOnlyTasksKey)
}

func loadMicrosoftVars() []string {
	microsoftVars := map[string]string{
		MicrosoftClientIDKey:     "",
		MicrosoftClientSecretKey: "",
		MicrosoftRedirectURLKey:  "",
	}

	unsetList := loadVarGroup(&microsoftVars, "microsoft")
	if len(unsetList) > 0 {
		return unsetList
	}

	MSClientID = microsoftVars[MicrosoftClientIDKey]
	MSClientSecret = microsoftVars[MicrosoftClientSecretKey]
	MSRedirectURL = microsoftVars[MicrosoftRedirectURLKey]

	return nil
}

func loadLambdaStoreVars() []string {
	lambdaStoreVars := map[string]string{
		LambdaHostKey:   "",
		LambdaSecretKey: "",
		LambdaPortKey:   "",
	}

	unsetList := loadVarGroup(&lambdaStoreVars, "lambda store")
	if len(unsetList) > 0 {
		return unsetList
	}

	LambdaHost = lambdaStoreVars[LambdaHostKey]
	LambdaPassword = lambdaStoreVars[LambdaSecretKey]
	LambdaPort = lambdaStoreVars[LambdaPortKey]

	return nil
}

func loadMongoDBVars() []string {
	mongoVars := map[string]string{
		MongoHostKey:     "",
		MongoUserKey:     "",
		MongoPasswordKey: "",
	}

	unsetList := loadVarGroup(&mongoVars, "mongo atlas")
	if len(unsetList) > 0 {
		return unsetList
	}

	MongoHost = mongoVars[MongoHostKey]
	MongoUser = mongoVars[MongoUserKey]
	MongoPassword = mongoVars[MongoPasswordKey]

	return nil
}

func loadVar(key string) (string, error) {
	variable := os.Getenv(key)
	if variable == "" {
		return "", errors.New(fmt.Sprintf("%q environment variable not set", key))
	}

	return variable, nil
}

func loadVarGroup(envKeys *map[string]string, groupName string) []string {
	localSlice := make([]string, 0)

	log.Printf("loading %v environment variables...\n", groupName)
	for key, _ := range *envKeys {
		v, err := loadVar(key)
		if err != nil {
			localSlice = append(localSlice, err.Error())
		} else {
			(*envKeys)[key] = v
		}
	}

	return localSlice
}

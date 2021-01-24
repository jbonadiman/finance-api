package redis

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"

	"github.com/jbonadiman/finance-bot/utils"
)

type DB utils.Connection

var (
	LambdaHost     string
	LambdaPassword string
	LambdaPort     string
)

func init() {
	var err error

	LambdaHost, err = utils.LoadVar("LAMBDA_HOST")
	if err != nil {
		log.Println(err.Error())
	}

	LambdaPassword, err = utils.LoadVar("LAMBDA_SECRET")
	if err != nil {
		log.Println(err.Error())
	}

	LambdaPort, err = utils.LoadVar("LAMBDA_PORT")
	if err != nil {
		log.Println(err.Error())
	}
}

func (db *DB) GetClient() (*redis.Client, error) {
	connectionStr := db.GetConnectionString()

	opt, err := redis.ParseURL(connectionStr)
	if err != nil {
		return nil, err
	}

	return redis.NewClient(opt), nil
}

func (db *DB) GetConnectionString() string {
	return fmt.Sprintf(
		"redis://%v:%v@%v:%v",
		db.User,
		db.Password,
		db.Host,
		db.Port)
}

func New() (*DB, error) {
	if LambdaHost == "" || LambdaPassword == "" || LambdaPort == "" {
		return nil, errors.New("lambda store credentials environment variables must be set")
	}

	db := DB{
		Host:     LambdaHost,
		Password: LambdaPassword,
		User:     "",
		Port:     LambdaPort,
	}

	return &db, nil
}

func GetTokenFromCache() (string, error) {
	log.Println("attempting to retrieve token from cache...")
	db, err := New()
	if err != nil {
		return "", err
	}

	redisClient, err := db.GetClient()
	if err != nil {
		return "", err
	}

	token := redisClient.Get(context.Background(), "token").Val()

	if token == "" {
		log.Println("token was not found on cache")
		return "", nil
	}

	log.Println("retrieved token from cache successfully")
	return token, nil
}

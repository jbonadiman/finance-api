package redis

import (
	"context"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"

	"github.com/jbonadiman/finance-bot/environment"
	"github.com/jbonadiman/finance-bot/utils"
)

type DB utils.Connection

var redisDB *DB

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

func GetDB() *DB {
	if redisDB == nil {
		redisDB = &DB{
			Host:     environment.LambdaHost,
			Password: environment.LambdaPassword,
			User:     "",
			Port:     environment.LambdaPort,
		}
	}

	return redisDB
}

func GetTokenFromCache() (string, error) {
	log.Println("attempting to retrieve token from cache...")

	redisClient, err := GetDB().GetClient()
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

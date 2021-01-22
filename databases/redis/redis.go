package redis

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/jbonadiman/finance-bot/databases"
	"github.com/jbonadiman/finance-bot/utils"
	"log"
)

type DB databases.Database

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

func New() *DB {
	lambdaHost, err := utils.LoadVar("LAMBDA_HOST")
	if err != nil {
		log.Println(err.Error())
	}

	lambdaPassword, err := utils.LoadVar("LAMBDA_SECRET")
	if err != nil {
		log.Println(err.Error())
	}

	lambdaPort, err := utils.LoadVar("LAMBDA_PORT")
	if err != nil {
		log.Println(err.Error())
	}

	db := DB{
		Host:     lambdaHost,
		Password: lambdaPassword,
		User:     "",
		Port:     lambdaPort,
	}

	return &db
}

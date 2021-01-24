package mongodb

import (
	"fmt"
	"github.com/jbonadiman/finance-bot/utils"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

type DB utils.Connection

func (db *DB) GetClient() (*mongo.Client, error) {
	connectionStr := db.GetConnectionString()

	client, err := mongo.NewClient(options.Client().ApplyURI(connectionStr))
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (db *DB) GetConnectionString() string {
	return fmt.Sprintf(
		"mongodb+srv://%v:%v@%v/finances?retryWrites=true&w=majoritycs",
		db.User,
		db.Password,
		db.Host)
}

func New() *DB {
	mongoHost, err := utils.LoadVar("MONGO_HOST")
	if err != nil {
		log.Println(err.Error())
	}

	mongoUser, err := utils.LoadVar("MONGO_USER")
	if err != nil {
		log.Println(err.Error())
	}

	mongoPassword, err := utils.LoadVar("MONGO_SECRET")
	if err != nil {
		log.Println(err.Error())
	}

	db := DB{
		Host:     mongoHost,
		Password: mongoPassword,
		User:     mongoUser,
		Port:     "",
	}

	return &db
}

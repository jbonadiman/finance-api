package mongodb

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/jbonadiman/finance-bot/entities"
	"github.com/jbonadiman/finance-bot/utils"
)

type DB struct {
	utils.Connection
	client         *mongo.Client
	IsDisconnected bool
}

const TimeOut = 5 * time.Second

var (
	MongoHost     string
	MongoPassword string
	MongoUser     string
)

func init() {
	var err error

	MongoHost, err = utils.LoadVar("MONGO_HOST")
	if err != nil {
		log.Println(err.Error())
	}

	MongoPassword, err = utils.LoadVar("MONGO_SECRET")
	if err != nil {
		log.Println(err.Error())
	}

	MongoUser, err = utils.LoadVar("MONGO_USER")
	if err != nil {
		log.Println(err.Error())
	}
}

func New() (*DB, error) {
	if MongoHost == "" || MongoPassword == "" || MongoUser == "" {
		return nil, errors.New("mongodb atlas credentials environment variables must be set")
	}

	db := DB{
		Connection: utils.Connection{
			Host:             MongoHost,
			Password:         MongoPassword,
			User:             MongoUser,
			Port:             "",
			ConnectionString: "",
		},
		IsDisconnected: true,
		client:         nil,
	}

	db.ConnectionString = fmt.Sprintf(
		"mongodb+srv://%v:%v@%v/finances?retryWrites=true&w=majority",
		db.User,
		db.Password,
		db.Host)

	client, err := mongo.NewClient(
		options.Client().ApplyURI(db.ConnectionString),
	)
	if err != nil {
		return nil, err
	}

	db.client = client

	ctx, cancel := context.WithTimeout(context.Background(), TimeOut)
	defer cancel()

	err = db.client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	db.IsDisconnected = false

	return &db, nil
}

func (db *DB) StoreTransactions(transactions ...entities.Transaction) (
	int,
	error,
) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeOut)
	defer cancel()

	if db.IsDisconnected {
		err := db.client.Connect(ctx)
		if err != nil {
			return 0, err
		}

		db.IsDisconnected = false
	}

	col := db.client.Database("finances").Collection("transactions")

	var items []interface{}
	for _, t := range transactions {
		items = append(items, t)
	}

	result, err := col.InsertMany(ctx, items)
	if err != nil {
		return 0, err
	}

	return len(result.InsertedIDs), nil
}

func (db *DB) GetCategory(unparsedCategory string) (*entities.Subcategory, error) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeOut)
	defer cancel()

	if db.IsDisconnected {
		err := db.client.Connect(ctx)
		if err != nil {
			return nil, err
		}

		db.IsDisconnected = false
	}

	col := db.client.Database("finances").Collection("subcategories")

	filter := bson.D{{"keywords", unparsedCategory}}

	sub := entities.Subcategory{}

	err := col.FindOne(ctx, filter).Decode(&sub)
	if err != nil {
		return nil, err
	}

	return &sub, nil
}

package mongodb

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/jbonadiman/finance-bot/entities"
	"github.com/jbonadiman/finance-bot/utils"
)

type DB struct {
	utils.Connection
	client                  *mongo.Client
	IsDisconnected          bool
	transactionsCollection  *mongo.Collection
	subcategoriesCollection *mongo.Collection
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
		db.Host,
	)

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

	financesDb := db.client.Database("finances")
	db.transactionsCollection = financesDb.Collection("transactions")
	db.subcategoriesCollection = financesDb.Collection("subcategories")

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

	var items []interface{}
	for _, t := range transactions {
		items = append(items, t)
	}

	result, err := db.transactionsCollection.InsertMany(ctx, items)
	if err != nil {
		return 0, err
	}

	return len(result.InsertedIDs), nil
}

func (db *DB) ParseCategory(unparsedCategory string) (
	*entities.Subcategory,
	error,
) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeOut)
	defer cancel()

	if db.IsDisconnected {
		err := db.client.Connect(ctx)
		if err != nil {
			return nil, err
		}

		db.IsDisconnected = false
	}

	filter := bson.D{{"keywords", unparsedCategory}}

	sub := entities.Subcategory{}

	err := db.subcategoriesCollection.FindOne(ctx, filter).Decode(&sub)
	if err != nil {
		return nil, err
	}

	return &sub, nil
}

func (db *DB) GetTransactionBySubcategory(subRegex string) (
	*[]entities.Transaction,
	error,
) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeOut)
	defer cancel()

	if db.IsDisconnected {
		err := db.client.Connect(ctx)
		if err != nil {
			return nil, err
		}

		db.IsDisconnected = false
	}

	filter := bson.M{"subcategory": primitive.Regex{Pattern: subRegex}}

	var transactions []entities.Transaction

	cursor, err := db.transactionsCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		currTransaction := entities.Transaction{}
		if err = cursor.Decode(&currTransaction); err != nil {
			log.Println(err.Error())
		}

		transactions = append(transactions, currTransaction)
	}

	log.Printf(
		"found %v transactions with the subcategory %q pattern",
		len(transactions),
		subRegex,
	)

	return &transactions, nil
}

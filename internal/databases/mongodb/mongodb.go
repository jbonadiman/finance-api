package mongodb

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/jbonadiman/finances-api/internal/entities"
	"github.com/jbonadiman/finances-api/internal/environment"
	"github.com/jbonadiman/finances-api/internal/utils"
)

type DB struct {
	utils.Connection
	client                  *mongo.Client
	IsDisconnected          bool
	transactionsCollection  *mongo.Collection
	subcategoriesCollection *mongo.Collection
}

const TimeOut = 5 * time.Second

var singleton *DB

func GetDB() (*DB, error) {
	if singleton == nil {
		singleton = &DB{
			Connection: utils.Connection{
				Host:             environment.MongoHost,
				Password:         environment.MongoPassword,
				User:             environment.MongoUser,
				Port:             "",
				ConnectionString: "",
			},
			IsDisconnected: true,
			client:         nil,
		}

		singleton.ConnectionString = fmt.Sprintf(
			"mongodb+srv://%v:%v@%v/finances?retryWrites=true&w=majority",
			singleton.User,
			singleton.Password,
			singleton.Host,
		)

		client, err := mongo.NewClient(
			options.Client().ApplyURI(singleton.ConnectionString),
		)

		if err != nil {
			return nil, err
		}

		singleton.client = client

		ctx, cancel := context.WithTimeout(context.Background(), TimeOut)
		defer cancel()

		err = singleton.client.Connect(ctx)
		if err != nil {
			return nil, err
		}

		singleton.IsDisconnected = false

		financesDb := singleton.client.Database("finances")
		singleton.transactionsCollection = financesDb.Collection("transactions")
		singleton.subcategoriesCollection = financesDb.Collection("subcategories")
	}

	return singleton, nil
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

func (db *DB) GetAllTransactions() (*[]entities.Transaction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeOut)
	defer cancel()

	if db.IsDisconnected {
		err := db.client.Connect(ctx)
		if err != nil {
			return nil, err
		}

		db.IsDisconnected = false
	}

	var transactions []entities.Transaction

	cursor, err := db.transactionsCollection.Find(ctx, bson.D{})
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
		"found %v transactions",
		len(transactions),
	)

	return &transactions, nil
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

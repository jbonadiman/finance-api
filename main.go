package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	handler "github.com/jbonadiman/finances-api/api"
	"github.com/jbonadiman/finances-api/entities"
)

func main() {
	http.HandleFunc(
		"/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("server is up"))
		},
	)

	http.HandleFunc("/api/auth", handler.StoreToken)
	http.HandleFunc("/api/get-tasks", handler.FetchTasks)
	http.HandleFunc("/api/query", handler.QueryTransactions)

	http.ListenAndServe(":8080", nil)
}

func loadSubcategories() {
	bgCtx := context.Background()

	mongoClient, err := mongo.NewClient(
		options.Client().ApplyURI(fmt.Sprintf(
			"mongodb+srv://%v:%v@%v/finances?retryWrites=true&w=majority",
			"dev-finances",
			"Gfr4fhUHPPilGQ46",
			"primary-cluster.o8pqa.mongodb.net",
		)),
	)

	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(bgCtx, 2 * time.Second)
	err = mongoClient.Connect(ctx)
	if err != nil {
		panic(err)
	}
	cancel()

	database := mongoClient.Database("finances")
	collection := database.Collection("subcategories")

	ctx, cancel = context.WithTimeout(bgCtx, 3 * time.Second)

	var subcategories []entities.Subcategory
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		panic(err)
	}

	for cursor.Next(ctx) {
		currSub := entities.Subcategory{}
		if err = cursor.Decode(&currSub); err != nil {
			panic(err)
		}

		subcategories = append(subcategories, currSub)
	}

	cursor.Close(ctx)
	cancel()

	opt, err := redis.ParseURL(
		"redis://:f75f28ea217738719f1dcbe72cd8a087@dory.redistogo.com:10826",
		// fmt.Sprintf(
		// "rediss://%v@%v:%v",
		// "3edaf75b3b4c46c89c5a934ef97ceb94",
		// "us1-loyal-aardvark-31728.lambda.store",
		// "31728"),
	)
	if err != nil {
		panic(err)
	}

	redisClient := redis.NewClient(opt)
	// ctx, cancel = context.WithTimeout(context.Background(), 1 * time.Second)
	// err = redisClient.Ping(ctx).Err()
	// if err != nil {
	// 	panic(err)
	// }
	cancel()

	keyValue := make([]string, 0)

	for _, subcategory := range subcategories {
		for _, keyword := range subcategory.Keywords {
			keyValue = append(keyValue, "subcategory:" + keyword, subcategory.Name)
		}
	}

	r := redisClient.MSet(context.Background(), keyValue)

	err = r.Err()
	if err != nil {
		panic(err)
	}
	cancel()
}

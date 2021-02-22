package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/oauth2"

	"github.com/jbonadiman/finances-api/internal/environment"
	"github.com/jbonadiman/finances-api/internal/utils"
)

type DB struct {
	utils.Connection
	client *redis.Client
}

const TimeOut = 3 * time.Second

var singleton *DB

func GetDB() (*DB, error) {
	if singleton == nil {
		singleton = &DB{
			Connection: utils.Connection{
				Host:     environment.LambdaHost,
				Password: environment.LambdaPassword,
				Port:     environment.LambdaPort,
			},
			client: nil,
		}
		singleton.ConnectionString = formatConnectionString(singleton)

		opt, err := redis.ParseURL(singleton.ConnectionString)
		if err != nil {
			return nil, err
		}

		client := redis.NewClient(opt)
		ctx, cancel := context.WithTimeout(context.Background(), TimeOut)
		defer cancel()

		err = client.Ping(ctx).Err()
		if err != nil {
			return nil, err
		}

		singleton.client = client
	}

	return singleton, nil
}

func formatConnectionString(db *DB) string {
	return fmt.Sprintf(
		"rediss://%v:%v@%v:%v",
		db.User,
		db.Password,
		db.Host,
		db.Port,
	)
}

func (db *DB) GetToken() (*oauth2.Token, error) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeOut)
	defer cancel()

	log.Println("attempting to retrieve token from cache...")

	values, err := db.client.MGet(ctx,
		"token:AccessToken",
		"token:RefreshToken",
		"token:TokenType",
		"token:Expiry").Result()

	if err != nil {
		return nil, err
	}

	accessToken := fmt.Sprint(values[0])
	refreshToken := fmt.Sprint(values[1])
	tokenType := fmt.Sprint(values[2])
	expiry, err := time.Parse(time.RFC3339Nano, fmt.Sprint(values[3]))

	if err != nil {
		return nil, err
	}

	if accessToken != "" && tokenType != "" && refreshToken != "" && !expiry.IsZero() {
		log.Println("retrieved token from cache successfully")

		token := oauth2.Token{
			AccessToken:  accessToken,
			TokenType:    tokenType,
			RefreshToken: refreshToken,
			Expiry:       expiry,
		}

		return &token, nil
	}

	return nil, nil
}

func (db *DB) StoreToken(token *oauth2.Token) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeOut)
	defer cancel()

	log.Println("storing token in cache...")
	db.client.MSet(ctx,
		"token:AccessToken", token.AccessToken,
		"token:RefreshToken", token.RefreshToken,
		"token:TokenType", token.TokenType,
		"token:Expiry", token.Expiry)
}

func (db *DB) CompareAuthentication(username, password string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), TimeOut)
	defer cancel()

	secret := db.client.Get(ctx, "auth:Secret").Val()
	return secret != "" && secret == username+":"+password
}

func (db *DB) ParseSubcategory(subcategory string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeOut)
	defer cancel()

	var parsedSubcategory, err = db.client.Get(
		ctx, fmt.Sprintf("subcategory:%v", subcategory)).Result()

	if err != nil {
		log.Printf("error parsing subcategory %q\n", subcategory)
		return "", err
	}

	return parsedSubcategory, nil
}

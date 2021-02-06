package redis

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/oauth2"

	"github.com/jbonadiman/finances-api/environment"
	"github.com/jbonadiman/finances-api/utils"
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

		singleton.client = redis.NewClient(opt)
	}

	return singleton, nil
}

func formatConnectionString(db *DB) string {
	return fmt.Sprintf(
		"redis://%v:%v@%v:%v",
		db.User,
		db.Password,
		db.Host,
		db.Port,
	)
}

func (db *DB) GetToken() (*oauth2.Token, error) {
	log.Println("attempting to retrieve token from cache...")
	wg := sync.WaitGroup{}

	var (
		accessToken,
		refreshToken,
		tokenType string
	)

	var expiry time.Time

	log.Println("getting token from cache...")

	wg.Add(4)
	go func() {
		defer wg.Done()
		accessToken = db.GetValue("token:AccessToken").Val()
	}()

	go func() {
		defer wg.Done()
		refreshToken = db.GetValue("token:RefreshToken").Val()
	}()

	go func() {
		defer wg.Done()
		tokenType = db.GetValue("token:TokenType").Val()
	}()

	go func() {
		defer wg.Done()
		var err error

		expiry, err = db.GetValue("token:Expiry").Time()
		if err != nil {
			log.Fatal(err.Error())
		}
	}()

	wg.Wait()

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
	log.Println("storing token in cache...")
	wg := sync.WaitGroup{}

	wg.Add(4)
	go func() {
		defer wg.Done()
		db.SetValue("token:AccessToken", token.AccessToken)
	}()

	go func() {
		defer wg.Done()
		db.SetValue("token:RefreshToken", token.RefreshToken)
	}()

	go func() {
		defer wg.Done()
		db.SetValue("token:TokenType", token.TokenType)
	}()

	go func() {
		defer wg.Done()
		db.SetValue("token:Expiry", token.Expiry)
	}()

	wg.Wait()
}

func (db *DB) CompareAuthentication(username, password string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), TimeOut)
	defer cancel()

	secret := db.client.Get(ctx, "auth:Secret").Val()
	return secret != "" && secret == username+":"+password
}

func (db *DB) GetValue(key string) *redis.StringCmd {
	ctx, cancel := context.WithTimeout(context.Background(), TimeOut)
	defer cancel()

	return db.client.Get(ctx, key)
}

func (db *DB) SetValue(key string, value interface{}) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeOut)
	defer cancel()

	db.client.Set(
		ctx,
		key,
		value,
		0,
	)
}

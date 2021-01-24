package utils

import (
	"context"
	"github.com/jbonadiman/finance-bot/databases/redis"
)

func GetTokenFromCache() (string, error) {
	db, err := redis.New()
	if err != nil {
		return "", err
	}

	redisClient, err := db.GetClient()
	if err != nil {
		return "", err
	}

	token := redisClient.Get(context.Background(), "token").Val()

	if token == "" {
		return "", nil
	}

	return token, nil
}

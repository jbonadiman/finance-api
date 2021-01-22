package databases

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/jbonadiman/finance-bot/utils"
	"log"
)

var (
	lambdaPassword string
	lambdaHost     string
	lambdaPort     string
)

func init() {
	var err error

	lambdaHost, err = utils.LoadVar("LAMBDA_HOST")
	if err != nil {
		log.Println(err.Error())
	}

	lambdaPassword, err = utils.LoadVar("LAMBDA_SECRET")
	if err != nil {
		log.Println(err.Error())
	}

	lambdaPort, err = utils.LoadVar("LAMBDA_PORT")
	if err != nil {
		log.Println(err.Error())
	}
}

func GetClient() *redis.Client {
	connectionStr := fmt.Sprintf("redis://:%v@%v:%v", lambdaPassword, lambdaHost, lambdaPort)

	opt, _ := redis.ParseURL(connectionStr)
	return redis.NewClient(opt)
}

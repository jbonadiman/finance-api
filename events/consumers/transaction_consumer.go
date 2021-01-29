package consumers

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"

	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/jbonadiman/finance-bot/brokers/rabbit"
	"github.com/jbonadiman/finance-bot/databases/mongodb"
	"github.com/jbonadiman/finance-bot/entities"

	"github.com/jbonadiman/finance-bot/models"
)

func init() {
	rabbitCon := rabbit.New()

	mongoClient, err := mongodb.New()
	if err != nil {
		log.Printf("could not connect to MongoDB: %v", err)
	} else {
		go func() {
			err = rabbitCon.RegisterConsumer(
				"tasksQueue",
				func(msg <-chan amqp.Delivery) {
					for event := range msg {
						log.Printf(
							"received an event of id %q\n",
							event.Timestamp,
						)
						task := models.Task{}

						err := json.Unmarshal(event.Body, &task)
						if err != nil {
							log.Printf(
								"unmarshalling of event %v resulted in the following error: %v\n",
								event,
								err,
							)
						}

						values := strings.Split(task.Title, ";")

						cost, err := strconv.ParseFloat(
							strings.TrimSpace(values[0]),
							64,
						)
						description := strings.TrimSpace(values[1])
						category := strings.TrimSpace(values[2])

						id, err := mongoClient.StoreOneTransaction(
							entities.Transaction{
								ID:             primitive.NewObjectID(),
								Date:           task.CreatedAt,
								CreatedAt:      task.CreatedAt,
								ModifiedAt:     task.ModifiedAt,
								OriginalTaskID: task.Id,
								Description:    description,
								Cost:           cost,
								Category:       category,
							},
						)
						if err != nil {
							log.Printf(
								"could not store the task of ID %v due to the error: %v\n",
								task.Id,
								err.Error(),
							)
						} else {
							log.Printf(
								"transaction %v created successfully!\n",
								id,
							)

							err = event.Ack(true)
							if err != nil {
								log.Printf(
									"could not acknowledge RabbitMQ event: %v\n",
									err,
								)
							}
						}
					}
				},
			)
			if err != nil {
				log.Printf("could not register consumer: %v\n", err)
			}
		}()
	}
}

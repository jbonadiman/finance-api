package rabbit

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/streadway/amqp"

	"github.com/jbonadiman/finance-bot/utils"
)

var (
	AmqpHost     string
	AmqpUser     string
	AmqpPassword string
)

type Rabbit utils.Connection

func init() {
	var err error

	AmqpHost, err = utils.LoadVar("CLOUDAMQP_HOST")
	if err != nil {
		log.Println(err.Error())
	}

	AmqpPassword, err = utils.LoadVar("CLOUDAMQP_SECRET")
	if err != nil {
		log.Println(err.Error())
	}

	AmqpUser, err = utils.LoadVar("CLOUDAMQP_USER")
	if err != nil {
		log.Println(err.Error())
	}
}

type ChannelReader interface {
	GetConnectionString() string
	RegisterConsumer(
		queueName string,
		action func(msg <-chan amqp.Delivery),
	) error
}

func (r *Rabbit) RegisterConsumer(
	queueName string,
	action func(msg <-chan amqp.Delivery),
) error {
	conn, err := amqp.Dial(r.GetConnectionString())
	if err != nil {
		log.Printf("Failed to connect to RabbitMQ: %v", err)
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Printf("Failed to open a channel: %v", err)
		return err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		log.Printf("Failed to declare a queue: %v", err)
		return err
	}

	taskEvents, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Printf("Failed to register a consumer: %s", err)
		return err
	}

	forever := make(chan bool)

	go action(taskEvents)

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever

	return nil
}

func (r *Rabbit) Publish(queue string, msg interface{}) error {
	conn, err := amqp.Dial(r.ConnectionString)
	if err != nil {
		log.Printf("failed to connect to RabbitMQ: %s", err)
		return err
	}

	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Printf("failed to open a channel: %s", err)
		return err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		queue, // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)

	if err != nil {
		log.Printf("failed to declare a queue: %s", err)
		return err
	}

	msgAsJson, err := json.Marshal(msg)
	if err != nil {
		log.Printf("failed to marshal the message: %s", err)
		return err
	}

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        msgAsJson,
		},
	)

	if err != nil {
		log.Printf("failed to publish a message: %s", err)
		return err
	}

	return nil
}

func (r *Rabbit) GetConnectionString() string {
	return fmt.Sprintf(
		"amqps://%v:%v@%v/%v",
		r.User,
		r.Password,
		r.Host,
		r.User)
}

func New() *Rabbit {
	r := &Rabbit{
		Host:     AmqpHost,
		Password: AmqpPassword,
		User:     AmqpUser,
		Port:     "",
	}

	r.ConnectionString = fmt.Sprintf(
		"amqps://%v:%v@%v/%v",
		r.User,
		r.Password,
		r.Host,
		r.User)

	return r
}

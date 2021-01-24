package rabbit

import (
	"fmt"
	"github.com/jbonadiman/finance-bot/utils"
	"github.com/streadway/amqp"
	"log"
)

var (
	AmqpHost string
	AmqpUser string
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
	RegisterConsumer(queueName string, action func(msg <-chan amqp.Delivery)) error
}

func (r *Rabbit) RegisterConsumer(queueName string, action func(msg <-chan amqp.Delivery)) error {
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
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
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

	go action(msgs)

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever

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

func New() (*Rabbit, error) {
	return &Rabbit{
		Host:     AmqpHost,
		Password: AmqpPassword,
		User:     AmqpUser,
		Port:     "",
	}, nil
}

package consumers

import (
	"github.com/streadway/amqp"
	"log"
)

func (c *Consumer) Consume(queue string, action func(msg <-chan amqp.Delivery)) error {
	conn, err := amqp.Dial(c.connString)
	if err != nil {
		log.Printf("Failed to connect to RabbitMQ: %s", err)
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Printf("Failed to open a channel: %s", err)
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
		log.Printf("Failed to declare a queue: %s", err)
		return err
	}

	msgs, err := ch.Consume(
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
package brokers

import "github.com/streadway/amqp"

type Consumer struct {
	Host string
	Password string
	User string
	Port string
	connectionString string
}

type ChannelReader interface {
	GetConnectionString() string
	Consume(queueName string, action func(msg <-chan amqp.Delivery)) error
}

func New(connection string) (*Consumer, error) {
	return &Consumer{connectionString: connString}, nil
}

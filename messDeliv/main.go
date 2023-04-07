package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Rabbit struct {
	Host       string
	Port       string
	VHost      string
	User       string
	Password   string
	NameQueue  string
	Heartbeat  int
	RabbitConn RabbitConn
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func (rabbit *Rabbit) fillingFieldsEnv() {
	rabbit.Host = rabbit.getEnvStr("ASD_RMQ_HOST", "localhost")
	rabbit.Port = rabbit.getEnvStr("ASD_RMQ_PORT", "5672")
	rabbit.VHost = rabbit.getEnvStr("ASD_RMQ_VHOST", "")
	rabbit.User = rabbit.getEnvStr("SEVICE_RMQ_ENOTIFY_USERNAME", "")
	rabbit.Password = rabbit.getEnvStr("SEVICE_RMQ_ENOTIFY_PASSWORD", "")
	rabbit.Heartbeat = rabbit.getEnvInt("ASD_RMQ_HEARTBEAT", 1)
	rabbit.NameQueue = rabbit.getEnvStr("ASD_RMQ_NAME_QUEUE", "")
}

func (rabbit Rabbit) getEnvStr(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func (rabbit Rabbit) getEnvInt(key string, defaultVal int) int {
	if value, exists := os.LookupEnv(key); exists {
		val, err := strconv.Atoi(value)
		if err != nil {
			log.Fatal(err)
		}
		return val
	}
	return defaultVal
}

func (rabbit *Rabbit) Connection() {
	loginParameters := fmt.Sprintf("amqp://%s:%s@%s:%s/%s", rabbit.User, rabbit.Password, rabbit.Host, rabbit.Port, rabbit.VHost)
	var err error
	rabbit.RabbitConn.Connector, err = amqp.Dial(loginParameters)
	failOnError(err, "Failed to connect to RabbitMQ")
	// defer conn.Close()

	rabbit.RabbitConn.Channel, err = rabbit.RabbitConn.Connector.Channel()
	failOnError(err, "Failed to open a channel")
	// defer ch.Close()

	if err = rabbit.RabbitConn.Channel.Qos(1, 0, false); err != nil {
		log.Fatal(err)
	}
}

func (rabbit *Rabbit) Consumer() (<-chan amqp.Delivery, error) {
	return rabbit.RabbitConn.Channel.Consume(
		rabbit.NameQueue,                      // queue
		"test_service",                        // consumer
		false,                                 // auto-ack
		false,                                 // exclusive
		false,                                 // no-local
		false,                                 // no-wait
		amqp.Table{"x-stream-offset": "last"}, // args
	)
}

type RabbitConn struct {
	Connector *amqp.Connection
	Channel   *amqp.Channel
}

func main() {
	configRabbit := Rabbit{}
	configRabbit.fillingFieldsEnv()
	configRabbit.Connection()

	m, err := configRabbit.Consumer()
	failOnError(err, "Failed to register a consumer")

	for d := range m {
		log.Printf("Received a message: %s", d.Body)
		d.Ack(true)
	}

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
}

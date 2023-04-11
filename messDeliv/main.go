package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	pgx "github.com/jackc/pgx/v5"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Rabbit struct {
	Host         string
	Port         string
	VHost        string
	User         string
	Password     string
	NameQueue    string
	Heartbeat    int
	RabbitConn   RabbitConn
	streamOffset int
}

type Postgres struct {
	Host         string
	Port         string
	User         string
	Password     string
	DataBaseName string
	Conn         *pgx.Conn
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func (pg *Postgres) pgEnv() {
	pg.Host = getEnvStr("ASD_POSTGRES_HOST", "localhost")
	pg.Port = getEnvStr("ASD_POSTGRES_PORT", "5432")
	pg.User = getEnvStr("ASD_POSTGRES_USER", "postgres")
	pg.Password = getEnvStr("ASD_POSTGRES_PASSWORD", "postgres")
	pg.DataBaseName = getEnvStr("ASD_POSTGRES_DBNAME", "postgres")
}

func (pg *Postgres) connPg() {
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", pg.User, pg.Password, pg.Host, pg.Port, pg.DataBaseName)
	var err error
	pg.Conn, err = pgx.Connect(context.Background(), dbURL)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func (pg *Postgres) requestDb(msg []byte, offset_msg int64) {
	_, err := pg.Conn.Exec(context.Background(), "call device.set_messages($1, $2)", msg, offset_msg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}
}

func (pg *Postgres) getOffset() int {
	var offset_msg int
	err := pg.Conn.QueryRow(context.Background(), "SELECT offset_msg FROM device.messages ORDER BY created_at DESC LIMIT 1;").Scan(&offset_msg)
	// offset_msg, err := pg.Conn.Exec(context.Background(), "select offset_msg from device.messages order by offset_msg limit 1;")
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
	}
	return offset_msg
}

func (rabbit *Rabbit) rabbitEnv() {
	rabbit.Host = getEnvStr("ASD_RMQ_HOST", "localhost")
	rabbit.Port = getEnvStr("ASD_RMQ_PORT", "5672")
	rabbit.VHost = getEnvStr("ASD_RMQ_VHOST", "")
	rabbit.User = getEnvStr("SEVICE_RMQ_ENOTIFY_USERNAME", "")
	rabbit.Password = getEnvStr("SEVICE_RMQ_ENOTIFY_PASSWORD", "")
	rabbit.Heartbeat = getEnvInt("ASD_RMQ_HEARTBEAT", 1)
	rabbit.NameQueue = getEnvStr("ASD_RMQ_NAME_QUEUE", "")
}

func getEnvStr(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if value, exists := os.LookupEnv(key); exists {
		val, err := strconv.Atoi(value)
		if err != nil {
			log.Fatal(err)
		}
		return val
	}
	return defaultVal
}

func (rabbit *Rabbit) connRabbit() {
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
	var args amqp.Table

	if rabbit.streamOffset > 0 {
		args = amqp.Table{"x-stream-offset": rabbit.streamOffset}
	} else {
		args = amqp.Table{"x-stream-offset": "last"}
	}

	return rabbit.RabbitConn.Channel.Consume(
		rabbit.NameQueue, // queue
		"test_service",   // consumer
		false,            // auto-ack
		false,            // exclusive
		false,            // no-local
		false,            // no-wait
		args,             // args
	)
}

type RabbitConn struct {
	Connector *amqp.Connection
	Channel   *amqp.Channel
}

func main() {
	confPg := Postgres{}
	configRabbit := Rabbit{}

	confPg.pgEnv()
	confPg.connPg()

	configRabbit.streamOffset = confPg.getOffset()
	if configRabbit.streamOffset > 0 {
		configRabbit.streamOffset += 1
	}

	// log.Printf("Received a message %s", offset_msg)

	configRabbit.rabbitEnv()
	configRabbit.connRabbit()

	m, err := configRabbit.Consumer()
	failOnError(err, "Failed to register a consumer")

	for d := range m {
		offset := d.Headers["x-stream-offset"].(int64)
		log.Printf("Received a message %d", offset)
		confPg.requestDb(d.Body, offset)
		log.Printf("Данные записаны в БД")
		d.Ack(true)
	}

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
}

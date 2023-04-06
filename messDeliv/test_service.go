package main

import (
	// "database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
)

// func failOnError(err error, msg string) {
// 	if err != nil {
// 		log.Panicf("%s: %s", msg, err)
// 	}
// }

func other() {

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	if err = ch.Qos(1, 0, false); err != nil {
		fmt.Println(err)
	}

	msgs, err := ch.Consume(
		"new",                                 // queue
		"consum",                              // consumer
		false,                                 // auto-ack
		false,                                 // exclusive
		false,                                 // no-local
		false,                                 // no-wait
		amqp.Table{"x-stream-offset": "last"}, // args
	)
	failOnError(err, "Failed to register a consumer")

	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			// message := d.Body
			// _, err := db.Exec("INSERT INTO test_schema.json_test(message) VALUES ($1);",
			// 	message)
			// failOnError(err, "Failed to open a channel")
			// defer db.Close()
			d.Ack(true)

		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

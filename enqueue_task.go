package main

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
)

// Submit this Task for completion
func EnqueueTask(typ string, params interface{}) error {
	if taskdefs[typ] == nil {
		return fmt.Errorf("unrecognized task type: '%s'", typ)
	}

	body, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("Error marshaling params to JSON: %s", err.Error())
	}

	// create the task locally to check validity
	// TODO - this should be moved into tasks package?
	t := taskdefs[typ]()
	if err := json.Unmarshal(body, t); err != nil {
		return fmt.Errorf("Error creating task from JSON: %s", err.Error())
	}

	if err := t.Valid(); err != nil {
		return fmt.Errorf("Invalid task: %s", err.Error())
	}

	conn, err := amqp.Dial(cfg.AmqpUrl)
	if err != nil {
		return fmt.Errorf("Failed to connect to RabbitMQ: %s", err.Error())
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("Failed to connect to open channel: %s", err.Error())
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"tasks", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		return fmt.Errorf("Failed to declare a queue: %s", err.Error())
	}

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Type:        typ,
			Body:        body,
		})

	if err != nil {
		return fmt.Errorf("Error publishing to queue: %s", err.Error())
	}

	return nil
}

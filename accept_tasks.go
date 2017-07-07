package main

import (
	"encoding/json"
	"fmt"
	"github.com/datatogether/task-mgmt/taskdefs/ipfs"
	"github.com/datatogether/task-mgmt/taskdefs/kiwix"
	"github.com/datatogether/task-mgmt/tasks"
	"github.com/streadway/amqp"
	"time"
)

// taskdefs is a map of all possible task names to their respective "New" funcs
var taskdefs = map[string]tasks.NewTaskFunc{
	"ipfs.add":            ipfs.NewTaskAdd,
	"kiwix.updateSources": kiwix.NewTaskUpdateSources,
}

// start accepting tasks from the queue, if setup doesn't error,
// it returns a stop channel writing to stop will teardown the
// func and stop accepting tasks
func acceptTasks() (stop chan bool, err error) {
	if cfg.AmqpUrl == "" {
		return nil, fmt.Errorf("no amqp url specified")
	}

	stop = make(chan bool)
	log.Printf("connecting to: %s", cfg.AmqpUrl)

	var conn *amqp.Connection
	for i := 0; i <= 1000; i++ {
		conn, err = amqp.Dial(cfg.AmqpUrl)
		if err != nil {
			log.Infof("Failed to connect to amqp server: %s", err.Error())
			time.Sleep(time.Second)
			continue
		}
		break
	}

	// TODO - handle connection still not existing

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("Failed to open a channel: %s", err.Error())
	}

	q, err := ch.QueueDeclare(
		"tasks", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("Error declaring que: %s", err.Error())
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return nil, fmt.Errorf("", err)
	}

	go func() {
		for msg := range msgs {
			// tasks.Tas
			task, err := tasks.TaskFromDelivery(store, msg)
			if err != nil {
				log.Errorf("task error: %s", err.Error())
				msg.Nack(false, false)
				continue
			}

			if err := task.Do(store); err != nil {
				log.Errorf("task error: %s", err.Error())
				msg.Nack(false, false)
			} else {
				log.Infof("completed task: %s, %s", msg.MessageId, msg.Type)
				msg.Ack(false)
			}

		}
		// TODO - figure out a way to bail out of the above loop
		// if stop is ever published to
		<-stop
		ch.Close()
		conn.Close()
	}()

	return stop, nil
}

// DoTask performs the designated task
func DoTask(msg amqp.Delivery) error {
	newTask := taskdefs[msg.Type]
	if newTask == nil {
		return fmt.Errorf("unknown task type: %s", msg.Type)
	}

	task := newTask()
	if err := json.Unmarshal(msg.Body, task); err != nil {
		return fmt.Errorf("error decoding task body json: %s", err.Error())
	}

	// If the task supports the DatastoreTask interface,
	// pass in our host db connection
	if dsT, ok := task.(tasks.DatastoreTaskable); ok {
		dsT.SetDatastore(store)
	}

	// created buffered progress updates channel
	pc := make(chan tasks.Progress, 10)

	// execute the task in a goroutine
	go task.Do(pc)

	for p := range pc {
		// TODO - log progress and pipe out of this func
		// so others can listen in for updates
		// log.Printf("")

		if p.Error != nil {
			return p.Error
		}
		if p.Done {
			return nil
		}
	}

	return nil
}

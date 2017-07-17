package main

import (
	"fmt"
	"github.com/datatogether/task-mgmt/taskdefs/ipfs"
	"github.com/datatogether/task-mgmt/taskdefs/kiwix"
	"github.com/datatogether/task-mgmt/tasks"
	"github.com/streadway/amqp"
	"time"
)

func configureTasks() {
	tasks.RegisterTaskdef("ipfs.addurl", ipfs.NewTaskAdd)
	tasks.RegisterTaskdef("ipfs.addcollection", ipfs.NewAddCollection)
	tasks.RegisterTaskdef("kiwix.updateSources", kiwix.NewTaskUpdateSources)

	// Must set api server url to make ipfs tasks work
	ipfs.IpfsApiServerUrl = cfg.IpfsApiUrl
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

	// if the connection is still nil after 1000 tries, time to bail
	if conn == nil {
		return nil, fmt.Errorf("Failed to connect to amqp server")
	}

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

			pch := make(chan tasks.Progress, 10)

			// accept tasks
			go func() {
				for {
					p := <-pch
					if err := PublishProgress(rconn, p); err != nil {
						log.Infoln(err.Error())
					}
				}
			}()

			if err := task.Do(store, pch); err != nil {
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

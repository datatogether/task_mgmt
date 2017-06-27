package main

import (
	"bytes"
	"fmt"
	"github.com/streadway/amqp"
	"io/ioutil"
	"mime/multipart"
	"net/http"
)

func acceptTasks() (stop chan bool, err error) {
	stop = make(chan bool)

	conn, err := amqp.Dial(cfg.AmqpUrl)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to RabbitMQ: %s", err.Error())
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("Failed to open a channel: %s", err.Error())
	}

	q, err := ch.QueueDeclare(
		"hello", // name
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
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return nil, fmt.Errorf("", err)
	}

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)

			body := &bytes.Buffer{}
			w := multipart.NewWriter(body)
			f, err := w.CreateFormFile("path", "hello-world.txt")
			if err != nil {
				log.Printf("error creating form file: %s", err.Error())
				continue
			}

			// TODO - handle errors
			f.Write([]byte("hello world"))

			if err := w.Close(); err != nil {
				log.Printf("error closing form file: %s", err.Error())
				continue
			}

			req, err := http.NewRequest("POST", fmt.Sprintf("%s/add", cfg.IpfsApiUrl), body)
			if err != nil {
				log.Printf("error creating request: %s", err.Error())
				continue
			}

			req.Header.Set("Content-Type", w.FormDataContentType())

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Printf("error sending request: %s", err.Error())
				continue
			}
			defer res.Body.Close()

			data, err := ioutil.ReadAll(res.Body)
			if err != nil {
				log.Printf("error reading response body: %s", err.Error())
				continue
			}
			log.Println(string(data))

		}
		<-stop
		ch.Close()
		conn.Close()
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")

	return stop, nil
}

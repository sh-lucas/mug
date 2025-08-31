package rabbit

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sh-lucas/mug/pkg"
)

var runningInTest bool
var rabbitUri string

func init() {
	if flag.Lookup("test.v") != nil {
		runningInTest = true
	}
	rabbitUri = os.Getenv("RABBIT_URI")

	if runningInTest {
		fmt.Println("Running in test mode. Messages will not be sent to RabbitMQ.")
	} else {
		if rabbitUri == "" {
			log.Fatalln(pkg.BoldRed + "You need to set the RABBIT_URI environment variable" + pkg.Reset)
		}
	}

	go func() {
		for connKeeper() {
			log.Println("Restarting connection keeper...")
			time.Sleep(1 * time.Second)
		}
	}()
}

type Conn struct {
	connection *amqp.Connection
	m          sync.RWMutex
}

var conn Conn

func connKeeper() (crash bool) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in connKeeper():", r)
			crash = true
		}
	}()

	if runningInTest {
		return
	}

	backoff := 200 * time.Millisecond
	const maxBackoff = 5 * time.Second

	for {
		conn.m.Lock()

		var err = fmt.Errorf("rabbitMQ is not ready yet")
		for err != nil {
			conn.connection, err = amqp.Dial(rabbitUri)
			if err != nil {
				time.Sleep(backoff)
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
				continue
			}
			backoff = 200 * time.Millisecond // reset backoff on success
		}

		conn.m.Unlock()

		watchedConn := conn.connection
		if watchedConn != nil {
			// waits for the connection to close
			err = <-watchedConn.NotifyClose(make(chan *amqp.Error))
		}

		log.Println("Connection closed:", err)
		time.Sleep(200 * time.Millisecond)
	}
}

var channels = make(chan *amqp.Channel, 50)

func Send(queue string, payload any) (ok bool) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in Send():", r)
			ok = false
		}
	}()

	body, err := jsoniter.Marshal(payload)
	if err != nil {
		log.Println("Failed to marshal payload:", err)
		return false
	}

	if runningInTest {
		log.Println("Test mode: Message not sent to RabbitMQ. Returning false from Send().")
		return false
	}

	select {
	case chann := <-channels:
		// gets a chann from the pool, verifies if it's still open
		if chann != nil && !chann.IsClosed() {
			ok = publish(chann, queue, body)
		} else {
			// reconnects if not
			if chann != nil {
				chann.Close()
			}
			chann = newChan()
			ok = publish(chann, queue, body)
		}
	default:
		chann := newChan()
		ok = publish(chann, queue, body)
	}

	return ok
}

// publish automatically returns the channel to the pool after publishing the message
// it returns false to any error, including a perfectly timed closed channel.
func publish(ch *amqp.Channel, queueName string, body []byte) (ok bool) {
	if ch.IsClosed() {
		ch = newChan()
	}
	log.Println("Publishing message to RabbitMQ:", body)

	// sends the payload and stuff
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Enable publisher confirms on the channel
	confirm, err := ch.PublishWithDeferredConfirmWithContext(
		ctx,
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        body,
		},
	)
	if err == nil && confirm.Wait() {
		// channel still alive <3
		select {
		case channels <- ch:
		default:
			ch.Close() // pool full, close the channel
		}
		return true
	} else {
		log.Printf("Failed to publish message, channel closed? error: %v", err)
		return false
	}
}

func newChan() (ch *amqp.Channel) {
	var err = fmt.Errorf("channel not created yet")

	for err != nil {
		conn.m.RLock()
		connection := conn.connection
		conn.m.RUnlock()

		if connection == nil || connection.IsClosed() {
			time.Sleep(200 * time.Millisecond)
			continue
		}

		ch, err = connection.Channel()
		if err != nil {
			time.Sleep(200 * time.Millisecond)
		}
	}

	// listen for channel close events
	go func(c *amqp.Channel) {
		defer func() {
			if r := recover(); r != nil {
				log.Println("Recovered in channel close listener:", r)
			}
		}()
		err := <-c.NotifyClose(make(chan *amqp.Error, 1))
		if err != nil {
			log.Printf("Channel closed: %v", err)
		}
	}(ch)

	return ch
}

// func Subscribe() {}

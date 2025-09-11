package rabbit

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sh-lucas/mug/pkg"
)

var rabbitUri string
var timeout = 30 * time.Second
var runningInTest bool

func detectRunningInTest() bool {
	if strings.HasSuffix(os.Args[0], ".test") {
		return true
	}
	for _, a := range os.Args[1:] {
		if strings.HasPrefix(a, "-test.") {
			return true
		}
	}
	return false
}

func init() {
	runningInTest = detectRunningInTest()
	go startup()
}

func startup() {
	// awaits a little for envs and stuff like that =)
	time.Sleep(100 * time.Millisecond)

	// setup timeout from env var
	timeoutStr := os.Getenv("RABBIT_TIMEOUT") // timeout for sending messages
	if timeoutStr != "" {
		t, err := time.ParseDuration(timeoutStr)
		if err != nil {
			panic("RABBIT_TIMEOUT environment variable is set but invalid: " + err.Error())
		} else {
			timeout = t
		}
	}

	// if flag.Lookup("test.v") != nil {
	// 	fmt.Println("Running in test mode. Messages will --- be sent to RabbitMQ.")
	// 	// runningInTest = true
	// 	return
	// }
	// rabbitUri = os.Getenv("RABBIT_URI")

	if rabbitUri == "" {
		log.Println(pkg.BoldRed + "You need to set the RABBIT_URI environment variable" + pkg.Reset)
	}

	go func() {
		for connKeeper() {
			log.Println("Panic catched, restarting connection keeper...")
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
		log.Println("Test mode: Message not sent to RabbitMQ. Returning true from Send().")
		return true
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
	log.Println("Publishing message to RabbitMQ:", string(body))

	// sends the payload and stuff
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
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
	if err == nil && confirm != nil && confirm.Wait() {
		// channel still alive <3
		select {
		case channels <- ch:
		default:
			ch.Close() // pool full, close the channel
		}
		return true
	} else {
		log.Printf("Failed to publish message, confirm=%v, error: %v", confirm, err)
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

		// set up confirm mode
		err = ch.Confirm(false)
		if err != nil {
			log.Printf("Failed to enable confirm mode: %v", err)
			ch.Close()
			time.Sleep(200 * time.Millisecond)
			continue // recreates the channel if something goes wrong
		}
	}

	// listen for channel close events
	go func(c *amqp.Channel) {
		// needs a recover if chann turns nil
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

// Ping checks if the RabbitMQ connection is alive by creating a temporary channel.
// Returns true if the connection is healthy without creating any queues or exchanges.
func Ping() bool {
	if runningInTest {
		log.Println("Test mode: Ping returning true (not actually checking RabbitMQ)")
		return true
	}

	conn.m.RLock()
	connection := conn.connection
	conn.m.RUnlock()

	if connection == nil || connection.IsClosed() {
		return false
	}

	// Create a temporary channel to test the connection
	ch, err := connection.Channel()
	if err != nil {
		return false
	}
	defer ch.Close()

	return !ch.IsClosed()
}

// Subscribe starts a pool of workers to process messages from the specified queue.
func Subscribe(queueName string, maxWorkers int, handler func(amqp.Delivery)) {
	if runningInTest {
		log.Printf("Test mode: WorkOnPool for queue '%s' not started", queueName)
		return
	}

	// limits the number of concurrent workers
	workerSem := make(chan struct{}, maxWorkers)

	log.Printf("Starting worker pool for queue '%s' with max %d workers", queueName, maxWorkers)

	// controls the spawn of workers
	go func() {
		for {
			workerSem <- struct{}{}
			go processWorker(queueName, workerSem, handler)
			time.Sleep(200 * time.Millisecond) // avoids hammering
		}
	}()
}

func processWorker(queueName string, workerSem <-chan struct{}, handler func(amqp.Delivery)) {
	defer func() { <-workerSem }() // release the semaphore when done
	defer func() {                 // recovers from panics just to be sure
		if r := recover(); r != nil {
			log.Printf("Worker recovered from panic: %v", r)
		}
	}()

	// owns it's own channel
	ch := newWorkerChannel()
	if ch == nil {
		log.Printf("Failed to create worker channel, exiting...")
		time.Sleep(2 * time.Second)
		return
	}
	defer ch.Close()

	_, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		log.Printf("Failed to declare queue '%s': %v", queueName, err)
		time.Sleep(2 * time.Second)
		return
	}

	// starts consuming with prefetch 5
	err = ch.Qos(5, 0, false)
	if err != nil {
		log.Printf("Failed to set QoS: %v", err)
		time.Sleep(2 * time.Second)
		return
	}

	// Consume
	msgs, err := ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		log.Printf("Failed to consume: %v", err)
		time.Sleep(2 * time.Second)
		return
	}

	log.Printf("Worker started consuming from '%s'", queueName)

	// Processa mensagens
	for msg := range msgs {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Handler panic: %v", r)
					msg.Nack(false, false) // nacks so it doesn't crash again
				}
			}()

			handler(msg)
		}()
	}

	log.Printf("Worker channel closed, will restart in 2s...")
	time.Sleep(2 * time.Second)
}

func newWorkerChannel() *amqp.Channel {
	for i := 0; i < 50; i++ { // Max 50 tentativas
		conn.m.RLock()
		connection := conn.connection
		conn.m.RUnlock()

		if connection == nil || connection.IsClosed() {
			backoff := time.Duration(i+1) * 200 * time.Millisecond
			if backoff > 5*time.Second {
				backoff = 5 * time.Second
			}
			time.Sleep(backoff)
			continue
		}

		ch, err := connection.Channel()
		if err != nil {
			backoff := time.Duration(i+1) * 200 * time.Millisecond
			if backoff > 5*time.Second {
				backoff = 5 * time.Second
			}
			time.Sleep(backoff)
			continue
		}

		return ch
	}
	return nil
}

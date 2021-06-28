package librabbitmq

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/helloferdie/stdgo/libstring"
	"github.com/helloferdie/stdgo/logger"

	"github.com/google/uuid"
	"github.com/streadway/amqp"
)

// Queue -
type Queue struct {
	Name             string
	Durable          bool
	Exclusive        bool
	DeleteWhenUnused bool
	NoWait           bool
}

// Resource -
type Resource struct {
	Queue        Queue
	errorChannel chan *amqp.Error
	connection   *amqp.Connection
	channel      *amqp.Channel
	closed       bool

	consumers []messageConsumer
}

// messageConsumer -
type messageConsumer func(amqp.Delivery)

// Payload -
type Payload struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}

// Producer -
type Producer struct {
	Resource  *Resource
	EventName string
	Payload   interface{}
}

// Connect - connect to rabbitmq server
func (r *Resource) Connect(log bool) error {
	conn, err := amqp.Dial(os.Getenv("amqp_connection"))
	if err != nil {
		if log {
			logger.MakeLogEntry(nil, false).Errorf("%s (%v)", "Failed to connect to RabbitMQ", err)
		}
		return err
	}
	r.connection = conn
	r.errorChannel = make(chan *amqp.Error)
	r.connection.NotifyClose(r.errorChannel)

	ch, err := r.connection.Channel()
	if err != nil {
		r.connection.Close()
		if log {
			logger.MakeLogEntry(nil, false).Errorf("%s (%v)", "Failed to open a channel", err)
		}
		return err
	}
	r.channel = ch

	_, err = r.channel.QueueDeclare(
		r.Queue.Name,             // name
		r.Queue.Durable,          // durable
		r.Queue.DeleteWhenUnused, // delete when unused
		r.Queue.Exclusive,        // exclusive
		r.Queue.NoWait,           // no-wait
		nil,                      // arguments
	)
	if err != nil {
		r.channel.Close()
		r.connection.Close()
		if log {
			logger.MakeLogEntry(nil, false).Errorf("%s (%v)", "Failed to declare a queue", err)
		}
		return err
	}
	if log {
		logger.PrintLogEntry("info", "Connected to message broker server", false)
	}
	return nil
}

// Reconnect - reconnect to rabbitmq server
func (r *Resource) Reconnect() {
	retryEnv := os.Getenv("amqp_reconnect_seconds")
	retry, err := strconv.Atoi(retryEnv)
	if err != nil {
		retry = 60
	}
	logger.PrintLogEntry("info", fmt.Sprintf("Consumer set to auto reconnect in %v seconds", retry), false)
	for {
		err := <-r.errorChannel
		if !r.closed {
			logger.PrintLogEntry("info", fmt.Sprintf("Connection lost %v", err), false)
			logger.PrintLogEntry("info", "Attempt to reconnect in "+retryEnv+" seconds", false)
			time.Sleep(time.Second * time.Duration(retry))
			err := r.Connect(true)
			if err == nil {
				r.recoverConsumers()
			}
		}
	}
}

// Close - close connection
func (r *Resource) Close() {
	//	log.Println("Closing connection")
	r.closed = true
	r.channel.Close()
	r.connection.Close()
}

// registerConsumer - register consumer
func (r *Resource) registerConsumer() (<-chan amqp.Delivery, error) {
	msgs, err := r.channel.Consume(
		r.Queue.Name, // queue
		"",           // consumer
		false,        // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	return msgs, err
}

// Consume - start consume instruction
func (r *Resource) Consume(consumer messageConsumer) error {
	logger.PrintLogEntry("info", "Register consumer ", false)
	msgs, err := r.registerConsumer()
	if err != nil {
		logger.PrintLogEntry("error", "Failed to register a consumer", false)
		return err
	}
	logger.PrintLogEntry("info", "Register consumer success, waiting message ", false)
	r.executeConsumer(err, consumer, msgs, false)
	return nil
}

// executeConsumer - listen and execute message from consumer
func (r *Resource) executeConsumer(err error, consumer messageConsumer, deliveries <-chan amqp.Delivery, isRecovery bool) {
	if err == nil {
		if !isRecovery {
			r.consumers = append(r.consumers, consumer)
		}
		go func() {
			for d := range deliveries {
				logger.PrintLogEntry("info", fmt.Sprintf("Receive message %v - %v", d.AppId, d.MessageId), false)
				consumer(d)
			}
		}()
	}
}

// recoverConsumers - recover dead consumer
func (r *Resource) recoverConsumers() {
	logger.PrintLogEntry("info", "Recovering consumer ", false)
	for c := range r.consumers {
		var consumer = r.consumers[c]
		msgs, err := r.registerConsumer()
		if err != nil {
			logger.PrintLogEntry("info", fmt.Sprintf("Failed to register consumer [%v]", c), false)
			continue
		}
		logger.PrintLogEntry("info", fmt.Sprintf("Register consumer [%v] success, processing message ", c), false)
		r.executeConsumer(err, consumer, msgs, true)
	}
	logger.PrintLogEntry("info", "Recovering consumer success", false)
}

// Publish - publish message
func Publish(r *Resource, appID string, payload map[string]interface{}) error {
	err := r.Connect(true)
	if err != nil {
		return err
	}

	nUUID := uuid.New()
	msgID := nUUID.String()
	err = r.channel.Publish(
		"",           // exchange
		r.Queue.Name, // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         []byte(libstring.JSONEncode(payload)),
			MessageId:    msgID,
			AppId:        appID,
			Headers: amqp.Table{
				"x-retry-count": 0,
			},
		})
	if err != nil {
		err = fmt.Errorf("%s", "Failed to publish a message")
		logger.MakeLogEntry(nil, false).Errorf("%v", err)
	}
	r.Close()
	logger.PrintLogEntry("info", "Message successfully publish "+appID+" "+msgID, false)
	return err
}

// Requeue - requeue message
func (r *Resource) Requeue(d amqp.Delivery) error {
	header := d.Headers
	retry := int(header["x-retry-count"].(int32))
	maxRetry, _ := strconv.Atoi(os.Getenv("amqp_max_retry"))
	if (maxRetry - 1) >= retry {
		retry++
		logger.PrintLogEntry("info", fmt.Sprintf("Requeue [%v] %s %s ", retry, d.AppId, d.MessageId), false)

		tmp := new(Resource)
		tmp.Queue = r.Queue
		err := tmp.Connect(false)
		if err != nil {
			return err
		}
		err = tmp.channel.Publish(
			"",             // exchange
			tmp.Queue.Name, // routing key
			false,          // mandatory
			false,          // immediate
			amqp.Publishing{
				DeliveryMode: d.DeliveryMode,
				ContentType:  d.ContentType,
				Body:         d.Body,
				MessageId:    d.MessageId,
				AppId:        d.AppId,
				Headers: amqp.Table{
					"x-retry-count": retry,
				},
			})
		tmp.Close()
		return err
	}
	r.Dump(d)
	logger.PrintLogEntry("info", fmt.Sprintf("Requeue reached max attempt [%v] %s %s ", retry, d.AppId, d.MessageId), false)
	return nil
}

// Dump - dump message to json file
func (r *Resource) Dump(d amqp.Delivery) {
	dir := os.Getenv("dir_dump") + "/" + d.AppId
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, 0777)
	}
	dir = dir + "/" + d.MessageId + ".json"
	destination, err := os.Create(dir)
	if err != nil {
		fmt.Println("os.Create:", err)
		return
	}
	defer destination.Close()
	os.Chmod(dir, 0777)
	fmt.Fprintf(destination, "%s", string(d.Body))
}

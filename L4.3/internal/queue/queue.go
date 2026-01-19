package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"task-manager/internal/logger"
	"task-manager/internal/models"

	rabbitmq "github.com/rabbitmq/amqp091-go"
)

type Queue struct {
	conn     *rabbitmq.Connection
	channel  *rabbitmq.Channel
	stopChan chan struct{}
	handler  func(context.Context, models.Notification) error
	wg       sync.WaitGroup
}

func NewQueue(url string) (*Queue, error) {
	conn, err := rabbitmq.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	q := &Queue{
		conn:     conn,
		channel:  ch,
		stopChan: make(chan struct{}),
	}

	// Настраиваем очереди
	if err := q.SetupQueues(); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}

	return q, nil
}

func (q *Queue) SetupQueues() error {
	// Создаём delayed exchange
	args := rabbitmq.Table{
		"x-delayed-type": "direct",
	}
	if err := q.channel.ExchangeDeclare(
		"notifications.delayed", // имя exchange
		"x-delayed-message",     // тип
		true,                    // durable
		false,                   // autoDelete
		false,                   // internal
		false,                   // noWait
		args,                    // arguments
	); err != nil {
		return err
	}

	// Создаём DLQ
	if _, err := q.channel.QueueDeclare(
		"notifications.dlq",
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   // args
	); err != nil {
		return err
	}

	// Создаём main очередь с DLQ
	mainQueueArgs := rabbitmq.Table{
		"x-dead-letter-exchange":    "",
		"x-dead-letter-routing-key": "notifications.dlq",
	}
	if _, err := q.channel.QueueDeclare(
		"notifications.main",
		true,
		false,
		false,
		false,
		mainQueueArgs,
	); err != nil {
		return err
	}

	// Привязываем main очередь к delayed exchange
	if err := q.channel.QueueBind(
		"notifications.main",    // queue
		"notifications.main",    // routing key
		"notifications.delayed", // exchange
		false,
		nil,
	); err != nil {
		return err
	}

	return nil
}

// Публикация
func (q *Queue) Publish(body []byte, sendAt interface{}) error {
	var delay time.Duration

	switch v := sendAt.(type) {
	case time.Time:
		delay = time.Until(v)
	case time.Duration:
		delay = v
	default:
		return fmt.Errorf("unsupported sendAt type: %T", sendAt)
	}

	if delay < 0 {
		delay = 0
	}

	err := q.channel.Publish(
		"notifications.delayed", // exchange
		"notifications.main",    // routing key
		false,                   // mandatory
		false,                   // immediate
		rabbitmq.Publishing{
			ContentType: "text/plain",
			Body:        body,
			Headers: rabbitmq.Table{
				"x-delay": int64(delay.Milliseconds()),
			},
		},
	)

	if err != nil {
		logger.Printf("Publish error: %v", err)
	} else {
		logger.Printf("Message published to exchange with delay %dms", delay.Milliseconds())
	}

	return err
}

// Запуск DLQ Consumer
func (q *Queue) StartDLQConsumer() {
	msgs, err := q.channel.Consume(
		"notifications.dlq",
		"",
		true,  // autoAck
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,   // args
	)
	if err != nil {
		logger.Println("DLQ consume error:", err)
		return
	}

	q.wg.Go(func() {
		for {
			select {
			case <-q.stopChan:
				logger.Println("DLQ consumer stopping...")
				return
			case msg, ok := <-msgs:
				if !ok {
					logger.Println("DLQ consumer channel closed")
					return
				}
				logger.Printf("DLQ сообщение: %s\n", string(msg.Body))
			}
		}
	})
}

// Запуск Main Consumer
func (q *Queue) StartMainConsumer() {
	msgs, err := q.channel.Consume(
		"notifications.main", // main очередь
		"",                   // consumer tag
		true,                 // autoAck
		false,                // exclusive
		false,                // noLocal
		false,                // noWait
		nil,                  // args
	)
	if err != nil {
		logger.Println("Main consume error:", err)
		return
	}

	q.wg.Go(func() {
		for {
			select {
			case <-q.stopChan:
				logger.Println("Main consumer stopping...")
				return
			case msg, ok := <-msgs:
				if !ok {
					logger.Println("Main consumer channel closed")
					return
				}
				var notification models.Notification
				if err := json.Unmarshal(msg.Body, &notification); err != nil {
					logger.Println("Failed to parse notification:", err)
					continue
				}

				if q.handler != nil {
					if err := q.handler(context.Background(), notification); err != nil {
						logger.Println("handler failed:", err)
					}
				} else {
					logger.Println("no handler set for queue")
				}
			}
		}
	})
}

func (q *Queue) SetHandler(handler func(context.Context, models.Notification) error) {
	q.handler = handler
}

// Close останавливает всех консьюмеров и закрывает соединения
func (q *Queue) Close() error {
	logger.Println("Closing queue consumers...")

	close(q.stopChan)
	q.wg.Wait()

	if err := q.channel.Close(); err != nil {
		return fmt.Errorf("failed to close channel: %w", err)
	}
	if err := q.conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}

	logger.Println("Queue closed.")
	return nil
}

package rabbitmq

import (
	"context"
	"fmt"
	"github.com/streadway/amqp"
	"log/slog"
	"time"
)

type BrokerConfig struct {
	URL           string
	ConsumeQueue  string
	PublishQueue  string
	PrefetchCount int
}

type BrokerRepoStruct struct {
	log          *slog.Logger
	conn         *amqp.Connection
	consumeCh    *amqp.Channel
	publishCh    *amqp.Channel
	consumeQueue amqp.Queue
	publishQueue amqp.Queue
}

func NewBrokerRepo(cfg BrokerConfig, log *slog.Logger) *BrokerRepoStruct {
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		panic(fmt.Sprintf("rabbitmq dial: %w", err))
	}

	consumeCh, err := conn.Channel()
	if err != nil {
		conn.Close()
		panic(fmt.Sprintf("open consume channel: %w", err))
	}

	publishCh, err := conn.Channel()
	if err != nil {
		consumeCh.Close()
		conn.Close()
		panic(fmt.Sprintf("open publish channel: %w", err))
	}

	cq, err := consumeCh.QueueDeclare(
		cfg.ConsumeQueue,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		consumeCh.Close()
		publishCh.Close()
		conn.Close()
		panic(fmt.Sprintf("declare consume queue: %w", err))
	}

	pq, err := publishCh.QueueDeclare(
		cfg.PublishQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		consumeCh.Close()
		publishCh.Close()
		conn.Close()
		panic(fmt.Sprintf("declare publish queue: %w", err))
	}

	if err := consumeCh.Qos(cfg.PrefetchCount, 0, false); err != nil {
		consumeCh.Close()
		publishCh.Close()
		conn.Close()
		panic(fmt.Sprintf("set QoS: %w", err))
	}

	return &BrokerRepoStruct{
		log:          log,
		conn:         conn,
		consumeCh:    consumeCh,
		publishCh:    publishCh,
		consumeQueue: cq,
		publishQueue: pq,
	}
}

func (b *BrokerRepoStruct) Close() error {
	b.consumeCh.Close()
	b.publishCh.Close()
	return b.conn.Close()
}

func (b *BrokerRepoStruct) ListenAndPublish(
	ctx context.Context,
	handler func(context.Context, []byte) ([]byte, error),
) error {
	op := "repository.ListenAndPublish"
	log := b.log.With(slog.String("op", op))
	msgs, err := b.consumeCh.Consume(
		b.consumeQueue.Name,
		"",
		false, // autoAck
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,
	)
	if err != nil {
		return fmt.Errorf("start consume: %w", err)
	}
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			select {
			case msg, ok := <-msgs:
				start := time.Now()

				if !ok {
					return fmt.Errorf("consume channel closed")
				}
				fmt.Println(string(msg.Body))
				if len(msg.Body) == 0 {
					log.Info("Received empty message, skipping...")
					continue
				}

				// Process the message
				taskCtx, cancel := context.WithCancel(ctx)
				result, err := handler(taskCtx, msg.Body)
				cancel()

				if err != nil {
					log.Error(fmt.Sprintf("handler error: %v, Nack and requeue", err))
					msg.Nack(false, false)
					continue
				}

				// Publish the result
				if err := b.publishCh.Publish(
					"", // default exchange
					b.publishQueue.Name,
					false, // mandatory
					false, // immediate
					amqp.Publishing{
						DeliveryMode: amqp.Persistent,
						ContentType:  "application/json",
						Body:         result,
					},
				); err != nil {
					log.Error(fmt.Sprintf("publish error: %v, Nack and requeue", err))
					msg.Nack(false, true)
					continue
				}

				msg.Ack(false)
				log.Info(fmt.Sprintf("%s", time.Since(start)))

			default:
				log.Info("No message, waiting for next tick...")
			}
		}
	}
}

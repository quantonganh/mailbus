package rabbitmq

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

type QueueService struct {
	ch *amqp.Channel
}

func NewQueueService(url string) (*QueueService, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &QueueService{
		ch: ch,
	}, nil
}

func (s *QueueService) Consume(ctx context.Context, topic string) (<-chan []byte, error) {
	q, err := s.ch.QueueDeclare(
		topic,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	deliveries, err := s.ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	messages := make(chan []byte)

	go func() {
		defer close(messages)

		for {
			select {
			case <-ctx.Done():
				return
			case d := <-deliveries:
				messages <- d.Body
			}
		}
	}()

	return messages, nil
}

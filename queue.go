package mailbus

import "context"

type QueueService interface {
	Consume(ctx context.Context, topic string) (<-chan []byte, error)
}

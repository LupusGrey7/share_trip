package kafka

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

// writerBatchTimeout caps how long a sync WriteMessages waits to fill a batch.
// kafka-go default is 1s: one event per call would hold the outbox TX ~1s per event.
const writerBatchTimeout = 10 * time.Millisecond

type TripEventProducer interface {
	PublishTripPublished(ctx context.Context, event TripPublished) error
}

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string, topic string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.Hash{},
			BatchTimeout: writerBatchTimeout,
		},
	}
}

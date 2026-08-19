package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type TripEventProducer interface {
	PublishTripPublished(ctx context.Context, event TripPublished) error
}

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string, topic string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.Hash{},
		},
	}
}

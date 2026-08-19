package kafka

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
)

func (p *Producer) PublishTripPublished(ctx context.Context, event TripPublished) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(
		ctx,
		kafka.Message{
			Key:   []byte(event.TripID), // key for massage
			Value: data,
		})
}

func (p *Producer) Close() error {
	return p.writer.Close()
}

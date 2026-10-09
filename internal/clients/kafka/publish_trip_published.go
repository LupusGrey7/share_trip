// scenario: publish trip published event to kafka

package kafka

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
)

func (p *Producer) PublishTripPublished(ctx context.Context, event TripPublished) error {
	data, err := json.Marshal(event) // validate event
	if err != nil {
		return err
	}

	// write message to kafka
	// key = trip_id (partitioning); value = envelope JSON
	return p.writer.WriteMessages( // returns error if message is not written
		ctx,
		kafka.Message{
			Key:   []byte(event.Payload.TripID),
			Value: data,
		})
}

func (p *Producer) Close() error {
	return p.writer.Close()
}

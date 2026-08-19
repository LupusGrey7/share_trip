// /scenario: publish trip published event to kafka
// /input: trip_id, driver_id, company_id
// /output: error if message is not written
// /error: if event is not valid
// /error: if message is not written
// /error: if writer is not closed
// /error: if writer is not closed
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
	// key for massage (trip_id is key) and value is event data
	return p.writer.WriteMessages( // returns error if message is not written
		ctx,
		kafka.Message{
			Key:   []byte(event.TripID),
			Value: data,
		})
}

func (p *Producer) Close() error {
	return p.writer.Close()
}

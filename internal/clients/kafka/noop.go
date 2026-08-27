package kafka

import "context"

// NoopProducer is used in apitest / when Kafka is not wired.
type NoopProducer struct{}

func (NoopProducer) PublishTripPublished(_ context.Context, _ TripPublished) error {
	return nil
}

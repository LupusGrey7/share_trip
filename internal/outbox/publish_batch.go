package outbox

import "context"

func (p *Publisher) publishBatch(ctx context.Context) error {
	events, err := p.outbox.LockPending(ctx, 100)
	if err != nil {
		return err
	}

	for _, event := range events {
		if err := p.kafka.Publish(ctx, event); err != nil {
			_ = p.outbox.MarkFailed(ctx, event.ID, err)
			continue
		}

		if err := p.outbox.MarkSent(ctx, event.ID); err != nil {
			return err
		}
	}

	return nil
}

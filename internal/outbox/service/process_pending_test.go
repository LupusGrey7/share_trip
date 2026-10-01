package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"job4j.ru/share_trip/internal/outbox/domain"
)

func TestSplitByHandle(t *testing.T) {
	t.Parallel()

	id1, id2, id3 := uuid.New(), uuid.New(), uuid.New()
	events := []*domain.Entity{{ID: id1}, {ID: id2}, {ID: id3}}
	brokerDown := errors.New("broker down")

	tests := []struct {
		name       string
		failIDs    map[uuid.UUID]bool
		wantSent   []uuid.UUID
		wantFailed []uuid.UUID
		wantErr    error
	}{
		{
			name:       "all sent",
			wantSent:   []uuid.UUID{id1, id2, id3},
			wantFailed: []uuid.UUID{},
		},
		{
			name:       "all failed",
			failIDs:    map[uuid.UUID]bool{id1: true, id2: true, id3: true},
			wantSent:   []uuid.UUID{},
			wantFailed: []uuid.UUID{id1, id2, id3},
			wantErr:    brokerDown,
		},
		{
			name:       "mixed keeps order",
			failIDs:    map[uuid.UUID]bool{id2: true},
			wantSent:   []uuid.UUID{id1, id3},
			wantFailed: []uuid.UUID{id2},
			wantErr:    brokerDown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var calls int
			handle := func(_ context.Context, e *domain.Entity) error {
				calls++
				if tt.failIDs[e.ID] {
					return brokerDown
				}
				return nil
			}

			sent, failed, lastErr := splitByHandle(context.Background(), events, handle)

			require.Equal(t, len(events), calls, "handle must be called for every event")
			require.Equal(t, tt.wantSent, sent)
			require.Equal(t, tt.wantFailed, failed)
			require.ErrorIs(t, lastErr, tt.wantErr)
		})
	}
}

func TestSplitByHandle_Empty(t *testing.T) {
	t.Parallel()

	sent, failed, lastErr := splitByHandle(context.Background(), nil, func(context.Context, *domain.Entity) error {
		t.Fatal("handle must not be called for empty batch")
		return nil
	})

	require.Empty(t, sent)
	require.Empty(t, failed)
	require.NoError(t, lastErr)
}

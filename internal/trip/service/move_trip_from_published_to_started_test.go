package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	contracts "job4j.ru/share_trip/internal/clients/http/contract"
	"job4j.ru/share_trip/internal/clients/http/contract/mocks"
	"job4j.ru/share_trip/internal/observability/metrics"
	"job4j.ru/share_trip/internal/trip/domain"
	"job4j.ru/share_trip/internal/trip/service"
	"job4j.ru/share_trip/internal/trip/usecase"
)

// fakeTripRepo returns trip for update and remembers a saved trip.
type fakeTripRepo struct {
	trip  *domain.Entity
	saved *domain.Entity
}

func (r *fakeTripRepo) UpdateTripTx(_ context.Context, _ pgx.Tx, t *domain.Entity) (*domain.Entity, error) {
	r.saved = t
	return t, nil
}

func (r *fakeTripRepo) GetForUpdateByIDTx(context.Context, pgx.Tx, string) (*domain.Entity, error) {
	return r.trip, nil
}

func (r *fakeTripRepo) GetTripByIDTx(context.Context, pgx.Tx, string) (*domain.Entity, error) {
	return nil, nil
}

func (r *fakeTripRepo) CreateTripDraftTx(context.Context, pgx.Tx, *domain.Entity) (*domain.Entity, error) {
	return nil, nil
}

// fakeTx implements only Commit and Rollback; other pgx.Tx methods are not called by tx().
type fakeTx struct {
	pgx.Tx
}

func (fakeTx) Commit(context.Context) error   { return nil }
func (fakeTx) Rollback(context.Context) error { return nil }

// fakeTxBeginner counts opened transactions.
type fakeTxBeginner struct {
	calls int
}

func (b *fakeTxBeginner) Begin(context.Context) (pgx.Tx, error) {
	b.calls++
	return fakeTx{}, nil
}

func newTripService(trips *fakeTripRepo, tx *fakeTxBeginner, contract contracts.ContractClient) *service.TripService {
	return service.NewTripService(
		metrics.New(prometheus.NewRegistry()),
		tx,
		trips,
		nil, // outbox repo
		usecase.NewTripUseCase(),
		contract,
	)
}

func startTripInput() domain.MoveTripFromPublishedToStartedInput {
	return domain.MoveTripFromPublishedToStartedInput{
		ID:          uuid.NewString(),
		ClientID:    uuid.New(),
		CompanyID:   "company-123",
		ServiceCode: domain.ServiceCodeTripStart,
	}
}

func publishedTrip(input domain.MoveTripFromPublishedToStartedInput) *domain.Entity {
	return &domain.Entity{
		ID:       uuid.MustParse(input.ID),
		DriverID: input.ClientID,
		Status:   domain.StatusPublished,
	}
}

// Scenario: Contract allows the service (allowed=true).
// Expect:   no error (HTTP 200), trip is started and saved, one transaction.
func TestTripService_MoveTripFromPublishedToStarted_Allowed(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	contract := mocks.NewMockContractClient(ctrl)
	contract.EXPECT().
		CheckAvailableService(gomock.Any(), contracts.CheckServiceRequest{CompanyID: "company-123", ServiceCode: "trip_start"}).
		Return(contracts.CheckResult{Allowed: true, Reason: "service_allowed"}, nil)

	input := startTripInput()
	trips := &fakeTripRepo{trip: publishedTrip(input)}
	tx := &fakeTxBeginner{}
	svc := newTripService(trips, tx, contract)

	res, err := svc.MoveTripFromPublishedToStarted(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, domain.StatusStarted, res.Status)
	assert.True(t, res.Allowed)
	assert.Equal(t, "service_allowed", res.Reason)
	require.NotNil(t, trips.saved)
	assert.Equal(t, domain.StatusStarted, trips.saved.Status)
	assert.Equal(t, 1, tx.calls)
}

// Scenario: Contract denies the service (allowed=false).
// Expect:   usecase.ErrConflict (HTTP 409), trip is not saved, no transaction.
func TestTripService_MoveTripFromPublishedToStarted_Denied(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	contract := mocks.NewMockContractClient(ctrl)
	contract.EXPECT().
		CheckAvailableService(gomock.Any(), contracts.CheckServiceRequest{CompanyID: "company-123", ServiceCode: "trip_start"}).
		Return(contracts.CheckResult{Allowed: false, Reason: "service_not_allowed"}, nil)

	input := startTripInput()
	trips := &fakeTripRepo{trip: publishedTrip(input)}
	tx := &fakeTxBeginner{}
	svc := newTripService(trips, tx, contract)

	res, err := svc.MoveTripFromPublishedToStarted(context.Background(), input)

	require.ErrorIs(t, err, usecase.ErrConflict)
	assert.Nil(t, res)
	assert.Nil(t, trips.saved)
	assert.Equal(t, 0, tx.calls)
}

// Scenario: Contract does not answer in time.
// Expect:   contracts.ErrTimeout (HTTP 504), trip is not saved, no transaction.
func TestTripService_MoveTripFromPublishedToStarted_Timeout(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	contract := mocks.NewMockContractClient(ctrl)
	contract.EXPECT().
		CheckAvailableService(gomock.Any(), contracts.CheckServiceRequest{CompanyID: "company-123", ServiceCode: "trip_start"}).
		Return(contracts.CheckResult{}, contracts.ErrTimeout)

	input := startTripInput()
	trips := &fakeTripRepo{trip: publishedTrip(input)}
	tx := &fakeTxBeginner{}
	svc := newTripService(trips, tx, contract)

	res, err := svc.MoveTripFromPublishedToStarted(context.Background(), input)

	require.ErrorIs(t, err, contracts.ErrTimeout)
	assert.Nil(t, res)
	assert.Nil(t, trips.saved)
	assert.Equal(t, 0, tx.calls)
}

// Scenario: Contract is down (5xx after retries).
// Expect:   contracts.ErrUnavailable (HTTP 503), trip is not saved, no transaction.
func TestTripService_MoveTripFromPublishedToStarted_Unavailable(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	contract := mocks.NewMockContractClient(ctrl)
	contract.EXPECT().
		CheckAvailableService(gomock.Any(), contracts.CheckServiceRequest{CompanyID: "company-123", ServiceCode: "trip_start"}).
		Return(contracts.CheckResult{}, contracts.ErrUnavailable)

	input := startTripInput()
	trips := &fakeTripRepo{trip: publishedTrip(input)}
	tx := &fakeTxBeginner{}
	svc := newTripService(trips, tx, contract)

	res, err := svc.MoveTripFromPublishedToStarted(context.Background(), input)

	require.ErrorIs(t, err, contracts.ErrUnavailable)
	assert.Nil(t, res)
	assert.Nil(t, trips.saved)
	assert.Equal(t, 0, tx.calls)
}

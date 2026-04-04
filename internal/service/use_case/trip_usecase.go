package use_case

import (
	"context"

	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/domain/trip"
	"job4j.ru/share_trip/internal/repository"
)

type BaseUsecase interface {
	CreateTrip(ctx context.Context, tx pgx.Tx, repo repository.BaseTxTripRepository, req trip.CreateTripRequest) (*trip.CreateTripResponse, error)
	MoveTripDraftToPublishTx(ctx context.Context, tx pgx.Tx, repo repository.BaseTxTripRepository, req trip.MoveTripDraftToPublishModel) (*trip.MoveTripDraftToPublishModelResponse, error)
}

type TripUsecase struct {
}

func NewTripUsecase() *TripUsecase {
	return &TripUsecase{}
}

//func (t *TripUsecase) CreateTrip(
//	ctx context.Context,
//	tx pgx.Tx,
//	repo repository.BaseTxTripRepository,
//	req trip.CreateTripRequest,
//) (*trip.CreateTripResponse, error) {
//	resp, err := CreateTrip(ctx, tx, repo, req)
//	if err != nil {
//		return nil, err
//	}
//	return resp, nil
//}

//func (t *TripUsecase) MoveTripDraftToPublishTx(
//	ctx context.Context,
//	tx pgx.Tx,
//	repo repository.BaseTxTripRepository,
//	req trip.MoveTripDraftToPublishModel,
//) (*trip.MoveTripDraftToPublishModelResponse, error) {
//	resp, err := MoveTripDraftToPublishTx(ctx, tx, repo, req)
//	if err != nil {
//		return nil, err
//	}
//	return resp, nil
//}

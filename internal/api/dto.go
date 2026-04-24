package api

import (
	"github.com/google/uuid"
	"job4j.ru/share_trip/internal/domain/trip"
)

type MoveTripDraftToPublishModelRequest struct {
	ID       string    `validate:"required,uuid"`
	ClientID uuid.UUID `json:"clientId" validate:"required,uuid"` //"omitempty,uuid"
}

func (req *MoveTripDraftToPublishModelRequest) ToRequest(id uuid.UUID) trip.MoveTripDraftToPublishModel {
	return trip.MoveTripDraftToPublishModel{
		ID:       id,
		ClientID: req.ClientID,
	}
}

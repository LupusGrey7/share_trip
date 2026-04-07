package api

import (
	"github.com/google/uuid"
	"job4j.ru/share_trip/internal/domain/trip"
)

type MoveTripDraftToPublishModelRequest struct {
	ID       uuid.UUID
	ClientID uuid.UUID `json:"clientId" validate:"required,uuid"` //"omitempty,uuid"
}

func (req *MoveTripDraftToPublishModelRequest) ToRequest() trip.MoveTripDraftToPublishModel {
	return trip.MoveTripDraftToPublishModel{
		ID:       req.ID,
		ClientID: req.ClientID,
	}
}

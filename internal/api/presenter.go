package api

import (
	"job4j.ru/share_trip/internal/trip/domain"
)

func toGetByIDInput(req *GetTripByIDRequest) *domain.GetByIDInput {
	if req == nil {
		return nil
	}
	return &domain.GetByIDInput{ID: req.ID}
}

func toGetTripByIDResponse(output *domain.GetTripByIDOutput) *GetTripByIDResponse {
	if output == nil {
		return nil
	}
	return &GetTripByIDResponse{
		ID:            output.ID,
		DriverID:      output.DriverID,
		FromPoint:     output.FromPoint,
		ToPoint:       output.ToPoint,
		Seats:         output.Seats,
		CreatedAt:     output.CreatedAt,
		DepartureTime: output.DepartureTime,
		Status:        StatusEnum(output.Status),
	}
}

// toCreateTripInput maps HTTP body → domain.
// DriverID is NOT taken from body — handler sets it from Keycloak sub.
func toCreateTripInput(req *CreateTripDraftRequest) *domain.CreateTripInput {
	if req == nil {
		return nil
	}
	return &domain.CreateTripInput{
		FromPoint:      req.FromPoint,
		ToPoint:        req.ToPoint,
		DepartureTime:  req.DepartureTime,
		AvailableSeats: req.AvailableSeats,
	}
}

func toCreateTripDraftResponse(output *domain.CreateTripOutput) *CreateTripDraftResponse {
	if output == nil {
		return nil
	}
	return &CreateTripDraftResponse{
		ID:            output.ID,
		DriverID:      output.DriverID,
		FromPoint:     output.FromPoint,
		ToPoint:       output.ToPoint,
		CreatedAt:     output.CreatedAt,
		DepartureTime: output.DepartureTime,
		Seats:         output.Seats,
		Status:        StatusEnum(output.Status),
	}
}

func toMoveTripDraftToPublishInput(req *MoveTripDraftToPublishRequest) domain.MoveTripDraftToPublishInput {
	if req == nil {
		return domain.MoveTripDraftToPublishInput{}
	}
	return domain.MoveTripDraftToPublishInput{
		ID:       req.ID,
		ClientID: req.ClientID,
	}
}

func toMoveTripDraftToPublishResponse(output *domain.MoveTripDraftToPublishOutput) *MoveTripDraftToPublishResponse {
	if output == nil {
		return nil
	}
	return &MoveTripDraftToPublishResponse{
		ID:            output.ID,
		DriverID:      output.DriverID,
		FromPoint:     output.FromPoint,
		ToPoint:       output.ToPoint,
		Seats:         output.Seats,
		CreatedAt:     output.CreatedAt,
		DepartureTime: output.DepartureTime,
		Status:        StatusEnum(output.Status),
	}
}

func toMoveTripPublishedToStartedInput(req MoveTripPublishedToStartedRequest) domain.MoveTripPublishedToStartedInput {
	return domain.MoveTripPublishedToStartedInput{
		ID:          req.ID,
		ClientID:    req.DriverID,
		CompanyID:   req.CompanyID,
		ServiceCode: domain.ServiceCodeEnum(req.ServiceCode),
	}
}

func toMoveTripPublishedToStartedResponse(
	output *domain.MoveTripPublishedToStartedOutput,
) *MoveTripPublishedToStartedResponse {
	if output == nil {
		return nil
	}
	return &MoveTripPublishedToStartedResponse{
		ID:            output.ID,
		DriverID:      output.DriverID,
		FromPoint:     output.FromPoint,
		ToPoint:       output.ToPoint,
		Seats:         output.Seats,
		CreatedAt:     output.CreatedAt,
		DepartureTime: output.DepartureTime,
		Status:        StatusEnum(output.Status),
		Allowed:       output.Allowed,
		Reason:        output.Reason,
	}
}

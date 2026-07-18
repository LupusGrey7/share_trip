package api

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"job4j.ru/share_trip/internal/api/apierr"
	"job4j.ru/share_trip/internal/domain/trip/usecase"
	"job4j.ru/share_trip/internal/middleware"
)

// helper function to check specific value in the chain of errors
func HandleError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, errors.New(apierr.ErrorClaimsNotFound)): //401
		return apierr.ErrResponse(c, fiber.StatusUnauthorized, apierr.ErrorClaimsNotFound) //401
	case errors.Is(err, apierr.ErrForbiddenRole): //403
		return apierr.ErrResponse(c, fiber.StatusForbidden, apierr.ErrorForbiddenRole) //403
	case errors.Is(err, usecase.ErrForbidden): //403
		return apierr.ErrResponse(c, fiber.StatusForbidden, errors.Unwrap(err).Error()) //  unwrap the error chain (extract the main description)
	case errors.Is(err, usecase.ErrTripNotFound): //404
		return apierr.ErrResponse(c, fiber.StatusNotFound, apierr.StatusNotFound) //404
	case errors.Is(err, usecase.ErrConflict):
		return apierr.ErrResponse(c, fiber.StatusConflict, errors.Unwrap(err).Error()) //409
	case errors.Is(err, apierr.ErrBadGateway):
		return apierr.ErrResponse(c, fiber.StatusBadGateway, apierr.ErrorBadGateway) //502
	default:
		return apierr.ErrResponse(c, fiber.StatusInternalServerError, apierr.InternalServerError) //500
	}
}

// get claims from context
// if token is missing, the application will return: 401 Unauthorized
// if token is valid, but the role is missing, the application will return: 403 Forbidden
func GetClaimsFromContext(c *fiber.Ctx) (*middleware.KeycloakClaims, error) {
	// 1. get interface{} from Map by key
	raw := c.Locals(middleware.KeycloakClaimsKey)

	// 2. make type assertion to *KeycloakClaims
	claims, ok := raw.(*middleware.KeycloakClaims)
	if !ok || claims == nil {
		// protection: if the middleware didn't work or the type didn't match
		return nil, errors.New(apierr.ErrorClaimsNotFound)
	}
	//check if the role is client
	if !claims.HasClientRole(middleware.KeycloakClientRole, "client") {
		return nil, apierr.ErrForbiddenRole
	}

	return claims, nil
}

// convert string to uuid.UUID
func convertStringToUUID(clientID string) (uuid.UUID, error) {
	clientIDUUID, err := uuid.Parse(clientID)
	if err != nil {
		return uuid.Nil, apierr.ErrBadGateway
	}
	return clientIDUUID, nil // UUID of the user in the Keycloak database
}

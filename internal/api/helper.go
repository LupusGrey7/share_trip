// package api — helper functions for the API layer
package api

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"job4j.ru/share_trip/internal/clients/http/contract"
	"job4j.ru/share_trip/internal/middleware"
	"job4j.ru/share_trip/internal/trip/usecase"
)

// HandleError maps domain/api sentinel errors to HTTP responses.
func HandleError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrClaimsNotFound): // 401
		return ErrResponse(c, fiber.StatusUnauthorized, ErrorClaimsNotFound)
	case errors.Is(err, ErrForbiddenRole): // 403
		return ErrResponse(c, fiber.StatusForbidden, ErrorForbiddenRole)
	case errors.Is(err, ErrForbiddenIDMismatch): // 403 — body driverId ≠ JWT sub
		return ErrResponse(c, fiber.StatusForbidden, ErrorForbiddenIDMismatch)
	case errors.Is(err, usecase.ErrForbidden): // 403
		if unwrapped := errors.Unwrap(err); unwrapped != nil {
			return ErrResponse(c, fiber.StatusForbidden, unwrapped.Error())
		}
		return ErrResponse(c, fiber.StatusForbidden, ErrorForbidden)
	case errors.Is(err, ErrInvalidValidate): // 400
		return ErrResponse(c, fiber.StatusBadRequest, RequestValidationError)
	case errors.Is(err, usecase.ErrTripNotFound): // 404
		return ErrResponse(c, fiber.StatusNotFound, StatusNotFound)
	case errors.Is(err, usecase.ErrConflict): // 409 - business deny (deny / wrong status / company not found)
		return ErrResponse(c, fiber.StatusConflict, err.Error())
	case errors.Is(err, contracts.ErrTimeout): // 504 — Contract timeout, fail closed
		return ErrResponse(c, fiber.StatusGatewayTimeout, ErrorContractTimeout)
	case errors.Is(err, contracts.ErrUnavailable): // 503 — Contract down / 5xx after retry
		return ErrResponse(c, fiber.StatusServiceUnavailable, ErrorContractUnavailable)
	case errors.Is(err, contracts.ErrBadRequest): // 400 — Contract отклонил запрос как невалидный
		return ErrResponse(c, fiber.StatusBadRequest, ErrorContractBadRequest)
	case errors.Is(err, contracts.ErrForbidden): // 403
		return ErrResponse(c, fiber.StatusForbidden, ErrorContractForbidden)
	case errors.Is(err, ErrBadGateway): // 502 — e.g. sub is not a UUID
		return ErrResponse(c, fiber.StatusBadGateway, ErrorBadGateway)
	default:
		return ErrResponse(c, fiber.StatusInternalServerError, InternalServerError)
	}
}

// GetClaimsFromContext returns JWT claims from Fiber locals (set by Keycloak middleware).
// 401 if claims missing; 403 if client role "client" is absent.
// Note: RequireClientRole on the route already checks the role — this is defense in depth.
func GetClaimsFromContext(c *fiber.Ctx) (*middleware.KeycloakClaims, error) {
	raw := c.Locals(middleware.KeycloakClaimsKey)

	claims, ok := raw.(*middleware.KeycloakClaims)
	if !ok || claims == nil {
		return nil, ErrClaimsNotFound
	}

	// HasClientRole(clientID, role) — first arg is OAuth client name, second is role
	if !claims.HasClientRole(middleware.KeycloakClientID, middleware.KeycloakClientRole) {
		return nil, ErrForbiddenRole
	}

	return claims, nil
}

// ClientIDFromClaims parses JWT subject (Keycloak user id) as UUID.
func ClientIDFromClaims(claims *middleware.KeycloakClaims) (uuid.UUID, error) {
	clientID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, ErrBadGateway
	}
	return clientID, nil
}

// package api — helper functions for the API layer
package api

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"job4j.ru/share_trip/internal/api/apierr"
	"job4j.ru/share_trip/internal/domain/trip/usecase"
	"job4j.ru/share_trip/internal/middleware"
)

// HandleError maps domain/api sentinel errors to HTTP responses.
func HandleError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, apierr.ErrClaimsNotFound): // 401
		return apierr.ErrResponse(c, fiber.StatusUnauthorized, apierr.ErrorClaimsNotFound)
	case errors.Is(err, apierr.ErrForbiddenRole): // 403
		return apierr.ErrResponse(c, fiber.StatusForbidden, apierr.ErrorForbiddenRole)
	case errors.Is(err, apierr.ErrForbiddenIDMismatch): // 403 — body driverId ≠ JWT sub
		return apierr.ErrResponse(c, fiber.StatusForbidden, apierr.ErrorForbiddenIDMismatch)
	case errors.Is(err, usecase.ErrForbidden): // 403
		if unwrapped := errors.Unwrap(err); unwrapped != nil {
			return apierr.ErrResponse(c, fiber.StatusForbidden, unwrapped.Error())
		}
		return apierr.ErrResponse(c, fiber.StatusForbidden, apierr.ErrorForbidden)
	case errors.Is(err, usecase.ErrTripNotFound): // 404
		return apierr.ErrResponse(c, fiber.StatusNotFound, apierr.StatusNotFound)
	case errors.Is(err, usecase.ErrConflict): // 409
		if unwrapped := errors.Unwrap(err); unwrapped != nil {
			return apierr.ErrResponse(c, fiber.StatusConflict, unwrapped.Error())
		}
		return apierr.ErrResponse(c, fiber.StatusConflict, "conflict")
	case errors.Is(err, apierr.ErrBadGateway): // 502 — e.g. sub is not a UUID
		return apierr.ErrResponse(c, fiber.StatusBadGateway, apierr.ErrorBadGateway)
	default:
		return apierr.ErrResponse(c, fiber.StatusInternalServerError, apierr.InternalServerError)
	}
}

// GetClaimsFromContext returns JWT claims from Fiber locals (set by Keycloak middleware).
// 401 if claims missing; 403 if client role "client" is absent.
// Note: RequireClientRole on the route already checks the role — this is defense in depth.
func GetClaimsFromContext(c *fiber.Ctx) (*middleware.KeycloakClaims, error) {
	raw := c.Locals(middleware.KeycloakClaimsKey)

	claims, ok := raw.(*middleware.KeycloakClaims)
	if !ok || claims == nil {
		return nil, apierr.ErrClaimsNotFound
	}

	// HasClientRole(clientID, role) — first arg is OAuth client name, second is role
	if !claims.HasClientRole(middleware.KeycloakClientID, middleware.KeycloakClientRole) {
		return nil, apierr.ErrForbiddenRole
	}

	return claims, nil
}

// ClientIDFromClaims parses JWT subject (Keycloak user id) as UUID.
func ClientIDFromClaims(claims *middleware.KeycloakClaims) (uuid.UUID, error) {
	clientID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, apierr.ErrBadGateway
	}
	return clientID, nil
}

package apitest

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"job4j.ru/share_trip/internal/middleware"
)

// testClientID — фиксированный sub для apitest (Keycloak stub).
var testClientID = uuid.MustParse("11111111-1111-1111-1111-111111111111")

// testKeycloakAuth injects JWT claims without Keycloak (apitest only).
func testKeycloakAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals(middleware.KeycloakClaimsKey, &middleware.KeycloakClaims{
			Subject: testClientID.String(),
			ResourceAccess: map[string]middleware.ResourceRoles{
				middleware.KeycloakClientID: {Roles: []string{middleware.KeycloakClientRole}},
			},
		})
		return c.Next()
	}
}

package fixtures

import (
	"sync"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"job4j.ru/share_trip/internal/middleware"
)

// Headers for requests (apitest stub does not call Keycloak; header is optional for realism).
var RefreshTokenHeader = "X-Refresh-Token"
var RefreshTokenValue = "refresh_token_xxxxx_01"

// InvalidClientID — другой user id (негативные ownership/IDOR кейсы).
var InvalidClientID = uuid.MustParse("11111111-1111-1111-1111-111111111112")

// NormalClientID — default stub sub (happy path).
var NormalClientID = uuid.MustParse("11111111-1111-1111-1111-111111111111")

// stubSubject — «кто сейчас залогинен» в тестовом middleware.
// TestMain регистрирует middleware один раз; тесты меняют только этот UUID.
var (
	stubMu           sync.Mutex
	stubSubject      = NormalClientID
	stubInjectClaims = true // false → Locals без claims → handler 401
)

// CurrentStubClientID returns the subject the stub will inject on the next request.
func CurrentStubClientID() uuid.UUID {
	stubMu.Lock()
	defer stubMu.Unlock()
	return stubSubject
}

// SetStubClientID switches the authenticated user for subsequent HTTP calls on testApp.
// Prefer UseStubClientID in tests — it restores the previous value via t.Cleanup.
func SetStubClientID(clientID uuid.UUID) {
	stubMu.Lock()
	defer stubMu.Unlock()
	stubSubject = clientID
}

// UseStubClientID sets stub subject for this test and restores the previous one after the test.
// Safe with shared TestMain app: we do not re-register routes / middleware.
func UseStubClientID(t *testing.T, clientID uuid.UUID) {
	t.Helper()
	stubMu.Lock()
	prev := stubSubject
	stubSubject = clientID
	stubMu.Unlock()

	t.Cleanup(func() {
		stubMu.Lock()
		stubSubject = prev
		stubMu.Unlock()
	})
}

// UseStubNoClaims — middleware не кладёт claims в Locals → GetClaimsFromContext → 401.
// uuid.Nil в UseStubClientID для этого НЕ подходит: stub всё равно создаёт claims с sub=Nil.
func UseStubNoClaims(t *testing.T) {
	t.Helper()
	stubMu.Lock()
	prev := stubInjectClaims
	stubInjectClaims = false
	stubMu.Unlock()

	t.Cleanup(func() {
		stubMu.Lock()
		stubInjectClaims = prev
		stubMu.Unlock()
	})
}

// ClaimsForClient builds claims like real JWT: sub + resource_access[share_trip-api]=[client].
func ClaimsForClient(clientID uuid.UUID) *middleware.KeycloakClaims {
	return &middleware.KeycloakClaims{
		Subject:           clientID.String(),
		PreferredUsername: "test",
		Email:             "test@test.com",
		AuthorizedParty:   middleware.KeycloakClientID,
		ResourceAccess: map[string]middleware.ResourceRoles{
			middleware.KeycloakClientID: {Roles: []string{middleware.KeycloakClientRole}},
		},
	}
}

// KeycloakRefreshTokenMiddleware — stub for TestMain: no Keycloak HTTP call.
// Each request reads current stubSubject (change via UseStubClientID / SetStubClientID).
func KeycloakRefreshTokenMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		stubMu.Lock()
		inject := stubInjectClaims
		subject := stubSubject
		stubMu.Unlock()

		if inject {
			c.Locals(middleware.KeycloakClaimsKey, ClaimsForClient(subject))
		}
		return c.Next()
	}
}

// Package middleware — Keycloak: refresh_token → access_token → claims в Fiber context.
// Учебный вариант: payload JWT читаем без проверки подписи (в prod — JWKS Keycloak).
package middleware

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	flog "github.com/gofiber/fiber/v2/log"
)

const (
	RefreshTokenHeader = "X-Refresh-Token" // RefreshTokenHeader — заголовок задания: Postman шлёт refresh_token сюда.
	KeycloakClaimsKey  = "keycloak_claims" // KeycloakClaimsKey — ключ в c.Locals(), куда middleware кладёт распарсенный JWT.
	KeycloakClientID   = "sharetrip-api"   // KeycloakClientID — имя OAuth-client в Keycloak (приложение sharetrip-api).
	KeycloakClientRole = "client"          // KeycloakClientRole — client role внутри sharetrip-api (задание: роль «client»).
)

type KeycloakConfig struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	HTTPClient   *http.Client
}

// KeycloakClaims — поля из payload access_token (JWT), которые нам нужны.
type KeycloakClaims struct {
	Subject           string `json:"sub"` // UUID пользователя в Keycloak
	PreferredUsername string `json:"preferred_username"`
	Email             string `json:"email"`
	AuthorizedParty   string `json:"azp"`
	ResourceAccess    map[string]struct {
		Roles []string `json:"roles"`
	} `json:"resource_access"`
}

type keycloakTokenResponse struct {
	AccessToken string `json:"access_token"`
}

// refreshAccessToken — обмен refresh_token на новый access_token через Keycloak token endpoint.
// Вызывается из middleware на каждый HTTP-запрос с X-Refresh-Token.
func refreshAccessToken(
	ctx context.Context,
	client *http.Client,
	cfg KeycloakConfig,
	refreshToken string,
) (*keycloakTokenResponse, error) {
	if cfg.Issuer == "" {
		return nil, errors.New("keycloak issuer is required")
	}
	if cfg.ClientID == "" {
		return nil, errors.New("keycloak client id is required")
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("client_id", cfg.ClientID)
	form.Set("refresh_token", refreshToken)

	if cfg.ClientSecret != "" {
		form.Set("client_secret", cfg.ClientSecret)
	}

	endpoint := strings.TrimRight(cfg.Issuer, "/") + "/protocol/openid-connect/token"

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		bytes.NewBufferString(form.Encode()),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("failed to close response body: %v", closeErr)
		}
	}()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("keycloak token endpoint returned status %d: %s", resp.StatusCode, string(body))
	}

	var token keycloakTokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("failed to unmarshal keycloak response: %v", err)
	}

	if token.AccessToken == "" {
		return nil, errors.New("keycloak response does not contain access_token")
	}

	return &token, nil
}

// parseAccessTokenClaims — декодирует среднюю часть JWT (payload) в KeycloakClaims.
// Подпись не проверяется — только учебное чтение claims.
func parseAccessTokenClaims(accessToken string) (*KeycloakClaims, error) {
	parts := strings.Split(accessToken, ".")
	if len(parts) != 3 {
		return nil, errors.New("jwt must contain three parts")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	var claims KeycloakClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, err
	}

	if claims.Subject == "" {
		return nil, errors.New("jwt does not contain sub claim")
	}

	return &claims, nil
}

// HasClientRole — this function is used to check if the client has the role, it is used to check if the client has the role.
func (c KeycloakClaims) HasClientRole(clientID string, role string) bool {
	access, ok := c.ResourceAccess[clientID]
	if !ok {
		return false
	}

	for _, current := range access.Roles {
		if current == role {
			return true
		}
	}

	return false
}

// RequireClientRole — this function is used to require the client role, it is used to require the client role.
// clientID — this is the client id, it is used to require the client role.
// role — this is the role, it is used to require the client role.
// Returns a fiber.Handler that requires the client role.
// If the token is missing, it returns a 401 Unauthorized error.
// If the client role is not present, it returns a 403 Forbidden error.
// If the client role is present, it returns a 200 OK error.
func RequireClientRole(clientID string, role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, err := ClaimsFromContext(c)
		if err != nil {
			return err
		}

		if !claims.HasClientRole(clientID, role) {
			return fiber.NewError(fiber.StatusForbidden, "access denied")
		}

		return c.Next()
	}
}

// ClaimsFromContext — this function is used to get the claims from the context, it is used to get the claims from the context.
func ClaimsFromContext(ctx *fiber.Ctx) (*KeycloakClaims, error) {
	value := ctx.Locals(KeycloakClaimsKey) // get the claims from the context

	claims, ok := value.(*KeycloakClaims) // cast the claims to the KeycloakClaims type
	if !ok {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "missing token claims") // return an error if the claims are not found
	}

	return claims, nil // return the claims
}

// KeycloakRefreshTokenMiddleware — основной middleware задания.
// 1) Читает X-Refresh-Token из заголовка.
// 2) POST в Keycloak → получает access_token.
// 3) Парсит claims → кладёт в Locals для RequireClientRole и handlers.
func KeycloakRefreshTokenMiddleware(cfg KeycloakConfig) fiber.Handler {
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}

	return func(c *fiber.Ctx) error {
		refreshToken := c.Get(RefreshTokenHeader)
		if refreshToken == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing refresh token")
		}

		token, err := refreshAccessToken(c.UserContext(), client, cfg, refreshToken)
		if err != nil {
			// fiber log — видно в make run (stdlib log мог теряться)
			flog.Errorf("keycloak refresh failed: %v", err)
			return fiber.NewError(fiber.StatusUnauthorized, "invalid refresh token")
		}

		claims, err := parseAccessTokenClaims(token.AccessToken)
		if err != nil {
			flog.Errorf("keycloak parse access token failed: %v", err)
			return fiber.NewError(fiber.StatusUnauthorized, "invalid access token")
		}

		c.Locals(KeycloakClaimsKey, claims)

		return c.Next()
	}
}

// Package middleware — Keycloak: refresh_token → access_token → claims в Fiber context.
// Educational variant: read payload JWT without signature verification (in prod — JWKS Keycloak).
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
	RefreshTokenHeader = "X-Refresh-Token" // RefreshTokenHeader — header of the task: Postman sends refresh_token here.
	KeycloakClaimsKey  = "keycloak_claims" // KeycloakClaimsKey — key in c.Locals(), where middleware puts the parsed JWT.
	KeycloakClientID   = "sharetrip-api"   // KeycloakClientID — name of the OAuth-client in Keycloak (application sharetrip-api).
	KeycloakClientRole = "client"          // KeycloakClientRole — client role inside sharetrip-api (task: role «client»).
)

type KeycloakConfig struct {
	Issuer       string       // Issuer — issuer of the token
	ClientID     string       // ClientID — client id of the token
	ClientSecret string       // ClientSecret — client secret of the token
	HTTPClient   *http.Client // HTTPClient — http client for the token
}

// ResourceAccess — resource access of the token
// ResourceRoles describes the list of roles for a specific resource
type ResourceRoles struct {
	Roles []string `json:"roles"` // Roles — roles of the user
}

// KeycloakClaims — fields from the payload of the access_token (JWT), that we need.
// Claims — this is the data that lies inside the JWT. We are interested in the sub field.
// `subject` field is the UUID of the user in the Keycloak database.
// We will use it as client_id in our business logic.
type KeycloakClaims struct {
	Subject           string                   `json:"sub"`                // Subject — subject of the token.UUID of the user in the Keycloak database.
	PreferredUsername string                   `json:"preferred_username"` // PreferredUsername — preferred username of the user
	Email             string                   `json:"email"`              // Email — email of the user
	AuthorizedParty   string                   `json:"azp"`                // AuthorizedParty — authorized party of the token
	ResourceAccess    map[string]ResourceRoles `json:"resource_access"`    // ResourceAccess — resource access of the token
}

type keycloakTokenResponse struct {
	AccessToken string `json:"access_token"` // AccessToken — access token of the token
}

// refreshAccessToken — exchange refresh_token for a new access_token through the Keycloak token endpoint.
// Call from middleware for each HTTP request with X-Refresh-Token.
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

	//var form url.Values //It is typically used for query parameters and form values. (map[string][]string)
	form := url.Values{} // initialize the form
	form.Set("grant_type", "refresh_token")
	form.Set("client_id", cfg.ClientID)
	form.Set("refresh_token", refreshToken)

	if cfg.ClientSecret != "" {
		form.Set("client_secret", cfg.ClientSecret)
	}

	endpoint := strings.TrimRight(cfg.Issuer, "/") + "/protocol/openid-connect/token"

	//NewRequestWithContext creates a new Request and assigns it to req for Keycloak token endpoint.
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

	resp, err := client.Do(req) //Do sends an HTTP request and returns an HTTP response, following policy (e.g. redirects, cookies, auth) for User-specified Host: Port.
	if err != nil {
		log.Printf("failed to do the request to the Keycloak token endpoint: %v", err)
		return nil, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("failed to close response body: %v", closeErr)
		}
	}()

	body, readErr := io.ReadAll(resp.Body) //var body []byte
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

// parseAccessTokenClaims — decodes the middle part of JWT (payload) into KeycloakClaims.
// The signature is not checked — only the claims are read for educational purposes.
func parseAccessTokenClaims(accessToken string) (*KeycloakClaims, error) {
	parts := strings.Split(accessToken, ".")
	if len(parts) != 3 {
		log.Printf("jwt must contain three parts")
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

// HasClientRole — this function is used TO CHECK IF THE CLIENT HAS THE ROLE, it is used to check if the client has the role.
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
func ClaimsFromContext(с *fiber.Ctx) (*KeycloakClaims, error) {
	value := с.Locals(KeycloakClaimsKey) // get the claims from the context

	claims, ok := value.(*KeycloakClaims) // cast the claims to the KeycloakClaims type
	if !ok {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "missing token claims") // return an error if the claims are not found
	}

	return claims, nil // return the claims
}

// KeycloakRefreshTokenMiddleware — the main middleware of the task.
// 1) Reads X-Refresh-Token from the header.
// 2) POST to Keycloak → receives access_token.
// 3) Parses claims → puts in Locals for RequireClientRole and handlers.
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
			// fiber log — visible in `make run` (stdlib log could be lost)
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

package middleware

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func KeycloakAuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Missing or invalid token"})
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			parts := strings.Split(token, ".")
			if len(parts) != 3 {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid JWT format"})
			}

			payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token payload"})
			}

			var claims map[string]interface{}
			if err := json.Unmarshal(payloadBytes, &claims); err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token json"})
			}

			// Store clientId (or whatever Keycloak maps to providerId) in context
			if clientId, ok := claims["clientId"].(string); ok {
				c.Set("providerId", clientId)
			} else if azp, ok := claims["azp"].(string); ok {
				c.Set("providerId", azp)
			}

			return next(c)
		}
	}
}

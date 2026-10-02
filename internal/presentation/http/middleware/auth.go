package middleware

import (
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

			// token := strings.TrimPrefix(authHeader, "Bearer ")
			// Here you would validate the JWT signature using github.com/golang-jwt/jwt
			// and Keycloak's public keys (JWKS).
			
			// For the challenge: extract ProviderID from claims
			// providerID := claims["clientId"].(string) 
			
			// Store in context to be validated against the payload later
			// c.Set("providerId", providerID)

			return next(c)
		}
	}
}

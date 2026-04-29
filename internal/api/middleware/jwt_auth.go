package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"DP_Maintenance/internal/model"
	"DP_Maintenance/internal/service"
)

type JWTMiddleware struct {
	authService *service.AuthService
}

func NewJWTMiddleware(authService *service.AuthService) *JWTMiddleware {
	return &JWTMiddleware{
		authService: authService,
	}
}

func (m *JWTMiddleware) Authenticate(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, model.ErrorResponse("missing authorization header"))
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.JSON(http.StatusUnauthorized, model.ErrorResponse("invalid authorization header format"))
		}

		tokenString := parts[1]
		claims, err := m.authService.ValidateToken(tokenString)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, model.ErrorResponse("invalid or expired token"))
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		return next(c)
	}
}

func GetUserID(c echo.Context) uint {
	if id, ok := c.Get("user_id").(uint); ok {
		return id
	}
	return 0
}

func GetUsername(c echo.Context) string {
	if username, ok := c.Get("username").(string); ok {
		return username
	}
	return ""
}

func GetRole(c echo.Context) string {
	if role, ok := c.Get("role").(string); ok {
		return role
	}
	return ""
}
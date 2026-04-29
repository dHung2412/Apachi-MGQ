package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"DP_Maintenance/internal/model"
	"DP_Maintenance/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
	users       map[string]*model.User
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		users:       make(map[string]*model.User),
	}
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req model.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse("invalid request body"))
	}

	if _, exists := h.users[req.Username]; exists {
		return c.JSON(http.StatusConflict, model.ErrorResponse("username already exists"))
	}

	hashedPassword, err := h.authService.HashPassword(req.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse("failed to hash password"))
	}

	user := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FullName:     req.FullName,
		Role:         "user",
		IsActive:     true,
	}
	h.users[user.Username] = user

	token, err := h.authService.GenerateToken(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse("failed to generate token"))
	}

	return c.JSON(http.StatusCreated, model.SuccessResponse(
		model.LoginResponse{Token: token, User: *user},
		"user registered successfully",
	))
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req model.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse("invalid request body"))
	}

	user, exists := h.users[req.Username]
	if !exists {
		return c.JSON(http.StatusUnauthorized, model.ErrorResponse("invalid credentials"))
	}

	if !h.authService.VerifyPassword(user.PasswordHash, req.Password) {
		return c.JSON(http.StatusUnauthorized, model.ErrorResponse("invalid credentials"))
	}

	token, err := h.authService.GenerateToken(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse("failed to generate token"))
	}

	return c.JSON(http.StatusOK, model.SuccessResponse(
		model.LoginResponse{Token: token, User: *user},
		"login successful",
	))
}
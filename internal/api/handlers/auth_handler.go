package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"DP_Maintenance/internal/model"
	"DP_Maintenance/internal/repository"
	"DP_Maintenance/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
	userRepo    repository.UserRepository
}

func NewAuthHandler(authService *service.AuthService, userRepo repository.UserRepository) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userRepo:    userRepo,
	}
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req model.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse("invalid request body"))
	}

	if h.userRepo != nil {
		existing, _ := h.userRepo.GetByUsername(req.Username)
		if existing != nil {
			return c.JSON(http.StatusConflict, model.ErrorResponse("username already exists"))
		}
	}

	hashedPassword, err := h.authService.HashPassword(req.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse("failed to hash password"))
	}

	user := &model.User{
		ID:           uuid.New(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FullName:     req.FullName,
		Role:         "user",
		IsActive:     true,
	}

	if h.userRepo != nil {
		if err := h.userRepo.Create(user); err != nil {
			return c.JSON(http.StatusInternalServerError, model.ErrorResponse("failed to create user"))
		}
	}

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

	var user *model.User
	var err error

	if h.userRepo != nil {
		user, err = h.userRepo.GetByUsername(req.Username)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, model.ErrorResponse("invalid credentials"))
		}
	}

	if user == nil || !h.authService.VerifyPassword(user.PasswordHash, req.Password) {
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
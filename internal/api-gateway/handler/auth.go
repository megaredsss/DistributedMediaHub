package handler

import (
	"fmt"
	"net/http"

	authpb "github.com/yourusername/DistributedMediaHub/api/proto/auth"
	"github.com/yourusername/DistributedMediaHub/internal/logger"
)

// AuthHandler обрабатывает запросы к Auth Service
type AuthHandler struct {
	authClient authpb.AuthServiceClient
	logger     *logger.Logger
}

// NewAuthHandler создает новый AuthHandler
func NewAuthHandler(authClient authpb.AuthServiceClient, log *logger.Logger) *AuthHandler {
	return &AuthHandler{
		authClient: authClient,
		logger:     log,
	}
}

// Register обрабатывает регистрацию нового пользователя
// POST /api/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req authpb.RegisterRequest
	if err := DecodeJSON(r, &req); err != nil {
		h.logger.WithError(err).Error("Failed to decode register request")
		WriteError(w, fmt.Errorf("invalid request body"), http.StatusBadRequest)
		return
	}

	// Валидация
	if req.Email == "" || req.Password == "" || req.Username == "" {
		WriteError(w, fmt.Errorf("email, password and username are required"), http.StatusBadRequest)
		return
	}

	// Вызываем Auth Service
	resp, err := h.authClient.Register(r.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Registration failed")
		WriteError(w, fmt.Errorf("registration failed"), http.StatusInternalServerError)
		return
	}

	h.logger.WithField("user_id", resp.UserId).Info("User registered successfully")
	WriteJSON(w, http.StatusCreated, resp)
}

// Login обрабатывает вход пользователя
// POST /api/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req authpb.LoginRequest
	if err := DecodeJSON(r, &req); err != nil {
		h.logger.WithError(err).Error("Failed to decode login request")
		WriteError(w, fmt.Errorf("invalid request body"), http.StatusBadRequest)
		return
	}

	// Валидация
	if req.Email == "" || req.Password == "" {
		WriteError(w, fmt.Errorf("email and password are required"), http.StatusBadRequest)
		return
	}

	// Вызываем Auth Service
	resp, err := h.authClient.Login(r.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Login failed")
		WriteError(w, fmt.Errorf("invalid credentials"), http.StatusUnauthorized)
		return
	}

	h.logger.WithField("user_id", resp.UserId).Info("User logged in successfully")
	WriteJSON(w, http.StatusOK, resp)
}

// Logout обрабатывает выход пользователя
// POST /api/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req authpb.LogoutRequest
	if err := DecodeJSON(r, &req); err != nil {
		h.logger.WithError(err).Error("Failed to decode logout request")
		WriteError(w, fmt.Errorf("invalid request body"), http.StatusBadRequest)
		return
	}

	// Вызываем Auth Service
	resp, err := h.authClient.Logout(r.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Logout failed")
		WriteError(w, fmt.Errorf("logout failed"), http.StatusInternalServerError)
		return
	}

	h.logger.WithField("user_id", req.UserId).Info("User logged out successfully")
	WriteJSON(w, http.StatusOK, resp)
}

// RefreshToken обрабатывает обновление токена
// POST /api/auth/refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req authpb.RefreshTokenRequest
	if err := DecodeJSON(r, &req); err != nil {
		h.logger.WithError(err).Error("Failed to decode refresh token request")
		WriteError(w, fmt.Errorf("invalid request body"), http.StatusBadRequest)
		return
	}

	// Вызываем Auth Service
	resp, err := h.authClient.RefreshToken(r.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Token refresh failed")
		WriteError(w, fmt.Errorf("token refresh failed"), http.StatusUnauthorized)
		return
	}

	h.logger.WithField("user_id", req.UserId).Info("Token refreshed successfully")
	WriteJSON(w, http.StatusOK, resp)
}

// GetUser обрабатывает получение информации о пользователе
// GET /api/auth/user/{id}
func (h *AuthHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	// TODO: Извлечь user_id из URL параметра
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		WriteError(w, fmt.Errorf("user_id is required"), http.StatusBadRequest)
		return
	}

	req := &authpb.GetUserRequest{
		UserId: userID,
	}

	// Вызываем Auth Service
	resp, err := h.authClient.GetUser(r.Context(), req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get user")
		WriteError(w, fmt.Errorf("user not found"), http.StatusNotFound)
		return
	}

	WriteJSON(w, http.StatusOK, resp)
}

// UpdateUser обрабатывает обновление информации о пользователе
// PUT /api/auth/user
func (h *AuthHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var req authpb.UpdateUserRequest
	if err := DecodeJSON(r, &req); err != nil {
		h.logger.WithError(err).Error("Failed to decode update user request")
		WriteError(w, fmt.Errorf("invalid request body"), http.StatusBadRequest)
		return
	}

	// Вызываем Auth Service
	resp, err := h.authClient.UpdateUser(r.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update user")
		WriteError(w, fmt.Errorf("update failed"), http.StatusInternalServerError)
		return
	}

	h.logger.WithField("user_id", req.UserId).Info("User updated successfully")
	WriteJSON(w, http.StatusOK, resp)
}

// DeleteUser обрабатывает удаление пользователя
// DELETE /api/auth/user/{id}
func (h *AuthHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		WriteError(w, fmt.Errorf("user_id is required"), http.StatusBadRequest)
		return
	}

	req := &authpb.DeleteUserRequest{
		UserId: userID,
	}

	// Вызываем Auth Service
	resp, err := h.authClient.DeleteUser(r.Context(), req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to delete user")
		WriteError(w, fmt.Errorf("delete failed"), http.StatusInternalServerError)
		return
	}

	h.logger.WithField("user_id", userID).Info("User deleted successfully")
	WriteJSON(w, http.StatusOK, resp)
}

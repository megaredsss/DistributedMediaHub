package auth

import (
	jwtAuth "github.com/yourusername/DistributedMediaHub/internal/auth-service/jwt"
	"github.com/yourusername/DistributedMediaHub/internal/logger"
)

type authService struct {
	logger     *logger.Logger
	jwtService jwtAuth.JWTService
	postgres   AuthRepository
}

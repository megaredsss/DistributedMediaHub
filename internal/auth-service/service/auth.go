package authservice

import (
	jwtAuth "github.com/yourusername/DistributedMediaHub/internal/auth-service/jwt"
	authrepository "github.com/yourusername/DistributedMediaHub/internal/auth-service/repository"
	"github.com/yourusername/DistributedMediaHub/internal/logger"
)

type authService struct {
	logger     *logger.Logger
	jwtService jwtAuth.JWTService
	postgres   authrepository.AuthRepository
}

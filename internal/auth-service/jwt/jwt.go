package jwtAuth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/yourusername/DistributedMediaHub/internal/logger"
)

// JWTService interface for JWT token management with security best practices
type JWTService interface {
	// GenerateAccessToken creates a short-lived access token
	GenerateAccessToken(ctx context.Context, userID, email string) (string, error)
	// GenerateRefreshToken creates a long-lived refresh token with unique JTI
	GenerateRefreshToken(ctx context.Context, userID, email string) (string, error)
	// ValidateAccessToken validates and parses access token, returns claims
	ValidateAccessToken(ctx context.Context, token string) (*AccessTokenClaims, error)
	// ValidateRefreshToken validates and parses refresh token, checks blacklist, returns claims
	ValidateRefreshToken(ctx context.Context, token string) (*RefreshTokenClaims, error)
	// RevokeToken adds refresh token to blacklist
	RevokeToken(ctx context.Context, token string) error
}

// AccessTokenClaims represents the claims for an access token
type AccessTokenClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Type   string `json:"typ"` // "access"
	jwt.RegisteredClaims
}

// RefreshTokenClaims represents the claims for a refresh token
type RefreshTokenClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Type   string `json:"typ"` // "refresh"
	jwt.RegisteredClaims
}

// TokenBlacklist defines interface for token revocation
type TokenBlacklist interface {
	Add(ctx context.Context, jti string, expiresAt time.Time) error
	IsBlacklisted(ctx context.Context, jti string) (bool, error)
}

type jwtService struct {
	accessSecret    []byte
	refreshSecret   []byte
	accessTokenTtl  time.Duration
	refreshTokenTtl time.Duration
	algorithm       jwt.SigningMethod
	issuer          string
	audience        []string
	logger          *logger.Logger
	blacklist       TokenBlacklist
}

// JWTServiceConfig holds configuration for JWT service
type JWTServiceConfig struct {
	AccessSecret    []byte
	RefreshSecret   []byte
	AccessTokenTtl  time.Duration
	RefreshTokenTtl time.Duration
	Algorithm       jwt.SigningMethod
	Issuer          string
	Audience        []string
	Logger          *logger.Logger
	Blacklist       TokenBlacklist // optional, can be nil
}

var (
	ErrInvalidTokenType = errors.New("invalid token type")
	ErrTokenRevoked     = errors.New("token has been revoked")
	ErrUnsafeAlgorithm  = errors.New("unsafe signing algorithm")
	ErrInvalidAlgorithm = errors.New("token uses different algorithm than expected")
	ErrEmptySecret      = errors.New("secret cannot be empty")
	ErrInvalidTTL       = errors.New("TTL must be positive")
)

// allowedAlgorithms contains whitelist of secure signing algorithms
var allowedAlgorithms = map[string]bool{
	"HS256": true, "HS384": true, "HS512": true,
	"RS256": true, "RS384": true, "RS512": true,
	"ES256": true, "ES384": true, "ES512": true,
}

// NewJwtService creates a new JWT service with security validations
func NewJwtService(cfg JWTServiceConfig) (JWTService, error) {
	// Validate algorithm
	if !allowedAlgorithms[cfg.Algorithm.Alg()] {
		return nil, fmt.Errorf("%w: %s", ErrUnsafeAlgorithm, cfg.Algorithm.Alg())
	}

	// Validate secrets
	if len(cfg.AccessSecret) == 0 || len(cfg.RefreshSecret) == 0 {
		return nil, ErrEmptySecret
	}

	// Validate TTL
	if cfg.AccessTokenTtl <= 0 || cfg.RefreshTokenTtl <= 0 {
		return nil, ErrInvalidTTL
	}

	// Set defaults
	if cfg.Issuer == "" {
		cfg.Issuer = "auth-service"
	}
	if len(cfg.Audience) == 0 {
		cfg.Audience = []string{"api-gateway"}
	}

	return &jwtService{
		accessSecret:    cfg.AccessSecret,
		refreshSecret:   cfg.RefreshSecret,
		accessTokenTtl:  cfg.AccessTokenTtl,
		refreshTokenTtl: cfg.RefreshTokenTtl,
		algorithm:       cfg.Algorithm,
		issuer:          cfg.Issuer,
		audience:        cfg.Audience,
		logger:          cfg.Logger,
		blacklist:       cfg.Blacklist,
	}, nil
}

func (s *jwtService) GenerateAccessToken(ctx context.Context, userID, email string) (string, error) {
	now := time.Now()

	claims := &AccessTokenClaims{
		UserID: userID,
		Email:  email,
		Type:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTokenTtl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.issuer,
			Audience:  s.audience,
		},
	}

	token := jwt.NewWithClaims(s.algorithm, claims)
	tokenString, err := token.SignedString(s.accessSecret)
	if err != nil {
		if s.logger != nil {
			s.logger.WithError(err).
				WithField("user_id", userID).
				Error("failed to sign access token")
		}
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}

	if s.logger != nil {
		s.logger.WithFields(map[string]interface{}{
			"user_id": userID,
			"ttl":     s.accessTokenTtl.String(),
		}).Info("access token generated")
	}

	return tokenString, nil
}

func (s *jwtService) GenerateRefreshToken(ctx context.Context, userID, email string) (string, error) {
	now := time.Now()
	jti := uuid.New().String() // Unique token ID for revocation

	claims := &RefreshTokenClaims{
		UserID: userID,
		Email:  email,
		Type:   "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(s.refreshTokenTtl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.issuer,
			Audience:  s.audience,
		},
	}

	token := jwt.NewWithClaims(s.algorithm, claims)
	tokenString, err := token.SignedString(s.refreshSecret)
	if err != nil {
		if s.logger != nil {
			s.logger.WithError(err).
				WithField("user_id", userID).
				Error("failed to sign refresh token")
		}
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	if s.logger != nil {
		s.logger.WithFields(map[string]interface{}{
			"user_id": userID,
			"jti":     jti,
			"ttl":     s.refreshTokenTtl.String(),
		}).Info("refresh token generated")
	}

	return tokenString, nil
}

func (s *jwtService) ValidateAccessToken(ctx context.Context, tokenString string) (*AccessTokenClaims, error) {
	claims := &AccessTokenClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			// Verify algorithm
			if token.Method.Alg() != s.algorithm.Alg() {
				if s.logger != nil {
					s.logger.WithFields(map[string]interface{}{
						"expected": s.algorithm.Alg(),
						"got":      token.Method.Alg(),
					}).Warn("invalid algorithm in access token")
				}
				return nil, ErrInvalidAlgorithm
			}
			return s.accessSecret, nil
		},
		jwt.WithLeeway(5*time.Second), // Clock skew tolerance
		jwt.WithValidMethods([]string{s.algorithm.Alg()}),
		jwt.WithIssuer(s.issuer),
		jwt.WithAudience(s.audience[0]),
	)

	if err != nil {
		if s.logger != nil {
			s.logger.WithError(err).Warn("access token validation failed")
		}
		return nil, fmt.Errorf("invalid access token: %w", err)
	}

	if !token.Valid {
		if s.logger != nil {
			s.logger.Warn("access token is not valid")
		}
		return nil, jwt.ErrTokenInvalidClaims
	}

	// Verify token type
	if claims.Type != "access" {
		if s.logger != nil {
			s.logger.WithField("type", claims.Type).
				Warn("wrong token type for access token")
		}
		return nil, ErrInvalidTokenType
	}

	if s.logger != nil {
		s.logger.WithField("user_id", claims.UserID).
			Debug("access token validated successfully")
	}

	return claims, nil
}

func (s *jwtService) ValidateRefreshToken(ctx context.Context, tokenString string) (*RefreshTokenClaims, error) {
	claims := &RefreshTokenClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			// Verify algorithm
			if token.Method.Alg() != s.algorithm.Alg() {
				if s.logger != nil {
					s.logger.WithFields(map[string]interface{}{
						"expected": s.algorithm.Alg(),
						"got":      token.Method.Alg(),
					}).Warn("invalid algorithm in refresh token")
				}
				return nil, ErrInvalidAlgorithm
			}
			return s.refreshSecret, nil
		},
		jwt.WithLeeway(5*time.Second), // Clock skew tolerance
		jwt.WithValidMethods([]string{s.algorithm.Alg()}),
		jwt.WithIssuer(s.issuer),
		jwt.WithAudience(s.audience[0]),
	)

	if err != nil {
		if s.logger != nil {
			s.logger.WithError(err).Warn("refresh token validation failed")
		}
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if !token.Valid {
		if s.logger != nil {
			s.logger.Warn("refresh token is not valid")
		}
		return nil, jwt.ErrTokenInvalidClaims
	}

	// Verify token type
	if claims.Type != "refresh" {
		if s.logger != nil {
			s.logger.WithField("type", claims.Type).
				Warn("wrong token type for refresh token")
		}
		return nil, ErrInvalidTokenType
	}

	// Check blacklist if available
	if s.blacklist != nil && claims.ID != "" {
		isBlacklisted, err := s.blacklist.IsBlacklisted(ctx, claims.ID)
		if err != nil {
			if s.logger != nil {
				s.logger.WithError(err).
					WithField("jti", claims.ID).
					Error("failed to check token blacklist")
			}
			return nil, fmt.Errorf("failed to check token blacklist: %w", err)
		}

		if isBlacklisted {
			if s.logger != nil {
				s.logger.WithFields(map[string]interface{}{
					"jti":     claims.ID,
					"user_id": claims.UserID,
				}).Warn("refresh token has been revoked")
			}
			return nil, ErrTokenRevoked
		}
	}

	if s.logger != nil {
		s.logger.WithFields(map[string]interface{}{
			"user_id": claims.UserID,
			"jti":     claims.ID,
		}).Debug("refresh token validated successfully")
	}

	return claims, nil
}

func (s *jwtService) RevokeToken(ctx context.Context, tokenString string) error {
	if s.blacklist == nil {
		return errors.New("token revocation is not supported: blacklist not configured")
	}

	// Parse token to get JTI and expiration
	claims := &RefreshTokenClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			if token.Method.Alg() != s.algorithm.Alg() {
				return nil, ErrInvalidAlgorithm
			}
			return s.refreshSecret, nil
		},
	)

	if err != nil {
		return fmt.Errorf("failed to parse token for revocation: %w", err)
	}

	if !token.Valid {
		return jwt.ErrTokenInvalidClaims
	}

	if claims.ID == "" {
		return errors.New("token does not have JTI, cannot revoke")
	}

	// Add to blacklist with expiration time
	expiresAt := claims.ExpiresAt.Time
	if err := s.blacklist.Add(ctx, claims.ID, expiresAt); err != nil {
		if s.logger != nil {
			s.logger.WithError(err).
				WithField("jti", claims.ID).
				Error("failed to add token to blacklist")
		}
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	if s.logger != nil {
		s.logger.WithFields(map[string]interface{}{
			"jti":     claims.ID,
			"user_id": claims.UserID,
		}).Info("token revoked successfully")
	}

	return nil
}

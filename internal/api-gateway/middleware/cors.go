package middleware

import (
	"net/http"
	"strings"

	"github.com/yourusername/DistributedMediaHub/internal/api-gateway/config"
)

// CORS middleware для обработки CORS запросов
type CORS struct {
	config *config.CORSConfig
}

// NewCORS создает новый CORS middleware
func NewCORS(cfg *config.CORSConfig) *CORS {
	return &CORS{
		config: cfg,
	}
}

// Middleware возвращает HTTP middleware для CORS
func (c *CORS) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Проверяем, разрешен ли origin
		if c.isOriginAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		// Устанавливаем разрешенные методы
		if len(c.config.AllowedMethods) > 0 {
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(c.config.AllowedMethods, ", "))
		}

		// Устанавливаем разрешенные заголовки
		if len(c.config.AllowedHeaders) > 0 {
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(c.config.AllowedHeaders, ", "))
		}

		// Устанавливаем Allow-Credentials
		if c.config.AllowCredentials {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// Обрабатываем preflight запросы
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// isOriginAllowed проверяет, разрешен ли origin
func (c *CORS) isOriginAllowed(origin string) bool {
	if origin == "" {
		return false
	}

	// Если список пустой, разрешаем все
	if len(c.config.AllowedOrigins) == 0 {
		return true
	}

	// Проверяем точное совпадение или wildcard
	for _, allowed := range c.config.AllowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}

		// Поддержка wildcard subdomain (например, *.example.com)
		if strings.HasPrefix(allowed, "*.") {
			domain := allowed[2:]
			if strings.HasSuffix(origin, domain) {
				return true
			}
		}
	}

	return false
}

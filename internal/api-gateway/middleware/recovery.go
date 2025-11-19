package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/yourusername/DistributedMediaHub/internal/logger"
)

// Recovery middleware для восстановления после паник
type Recovery struct {
	logger *logger.Logger
}

// NewRecovery создает новый Recovery middleware
func NewRecovery(log *logger.Logger) *Recovery {
	return &Recovery{
		logger: log,
	}
}

// Middleware возвращает HTTP middleware для recovery
func (rec *Recovery) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Логируем panic с stack trace
				rec.logger.WithFields(map[string]interface{}{
					"panic":      err,
					"stacktrace": string(debug.Stack()),
					"method":     r.Method,
					"path":       r.URL.Path,
					"client_ip":  getClientIP(r),
				}).Error("Panic recovered")

				// Отправляем Internal Server Error
				http.Error(w, fmt.Sprintf("Internal Server Error: %v", err), http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

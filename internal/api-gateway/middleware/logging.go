package middleware

import (
	"net/http"
	"time"

	"github.com/yourusername/DistributedMediaHub/internal/logger"
	"github.com/go-chi/chi/v5/middleware"
)

// Logging middleware для логирования HTTP запросов
type Logging struct {
	logger *logger.Logger
}

// NewLogging создает новый Logging middleware
func NewLogging(log *logger.Logger) *Logging {
	return &Logging{
		logger: log,
	}
}

// Middleware возвращает HTTP middleware для логирования
func (l *Logging) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Оборачиваем ResponseWriter для получения статус кода
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		// Выполняем запрос
		next.ServeHTTP(ww, r)

		// Вычисляем длительность
		duration := time.Since(start)

		// Логируем запрос
		l.logger.WithFields(map[string]interface{}{
			"method":      r.Method,
			"path":        r.URL.Path,
			"status":      ww.Status(),
			"duration_ms": duration.Milliseconds(),
			"client_ip":   getClientIP(r),
			"user_agent":  r.UserAgent(),
			"bytes":       ww.BytesWritten(),
		}).Info("HTTP request")
	})
}

// getClientIP извлекает IP адрес клиента
func getClientIP(r *http.Request) string {
	// Проверяем заголовки прокси
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}

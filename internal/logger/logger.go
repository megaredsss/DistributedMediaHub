package logger

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Package logger provides a thin wrapper around zerolog with convenience helpers
// for structured logging, context propagation, and common request logging
// patterns (HTTP and gRPC). It centralizes service and environment fields and
// exposes small helpers to create preconfigured loggers for different
// environments.

type Environment string

const (
	// EnvironmentDev represents the development environment.
	EnvironmentDev Environment = "development"
	// EnvironmentStage represents the staging environment.
	EnvironmentStage Environment = "staging"
	// EnvironmentProd represents the production environment.
	EnvironmentProd Environment = "production"
)

// Config holds configuration for building a logger.
type Config struct {
	// ServiceName is added to every log entry as the "service" field.
	ServiceName string
	// Environment is added to every log entry as the "environment" field.
	Environment Environment
	// Level defines the minimum log level (e.g. "debug", "info", "error").
	Level string
	// Pretty toggles console pretty output (ConsoleWriter) vs JSON.
	Pretty bool
}

// Logger is a thin wrapper around a zerolog.Logger pointer that provides
// convenience helpers for structured fields, context propagation, and
// request-specific logging.
type Logger struct {
	logger *zerolog.Logger
}

type contextKey string

const loggerCtxKey contextKey = "logger"

// New constructs a Logger from a Config. It sets the output writer, log level,
// timestamps, and attaches service/environment fields.
func New(cfg Config) *Logger {
	var writer io.Writer
	if cfg.Pretty {
		writer = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	} else {
		writer = os.Stdout
	}

	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}

	zlog := zerolog.New(writer).
		Level(level).
		With().
		Timestamp().
		Str("service", cfg.ServiceName).
		Str("environment", string(cfg.Environment)).
		Logger()

	return &Logger{logger: &zlog}
}

// CreateLogger returns a new Logger. Kept for compatibility with factories
// that expect a constructor that returns an error.
func CreateLogger(cfg Config) (*Logger, error) {
	return New(cfg), nil
}

// With returns a new Logger that shares the underlying zerolog but has a new
// With() context attached (useful to build up structured fields).
func (l *Logger) With() *Logger {
	newLogger := l.logger.With().Logger()
	return &Logger{logger: &newLogger}
}

// WithField returns a new Logger with a single structured field attached.
func (l *Logger) WithField(key string, value interface{}) *Logger {
	newLogger := l.logger.With().Interface(key, value).Logger()
	return &Logger{logger: &newLogger}
}

// WithFields returns a new Logger with multiple structured fields attached.
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	ctx := l.logger.With()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	newLogger := ctx.Logger()
	return &Logger{logger: &newLogger}
}

// WithError returns a new Logger with an error attached as the "error" field.
func (l *Logger) WithError(err error) *Logger {
	if err == nil {
		return l
	}
	newLogger := l.logger.With().Err(err).Logger()
	return &Logger{logger: &newLogger}
}

// WithTraceID returns a new Logger with a trace_id field attached.
func (l *Logger) WithTraceID(traceID string) *Logger {
	newLogger := l.logger.With().Str("trace_id", traceID).Logger()
	return &Logger{logger: &newLogger}
}

// WithRequestID returns a new Logger with a request_id field attached.
func (l *Logger) WithRequestID(requestID string) *Logger {
	newLogger := l.logger.With().Str("request_id", requestID).Logger()
	return &Logger{logger: &newLogger}
}

// WithUserID returns a new Logger with a user_id field attached.
func (l *Logger) WithUserID(userID string) *Logger {
	newLogger := l.logger.With().Str("user_id", userID).Logger()
	return &Logger{logger: &newLogger}
}

// ToContext stores the Logger in the provided context and returns the new ctx.
func (l *Logger) ToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, loggerCtxKey, l)
}

// FromContext retrieves a Logger from the context. If none is present, returns
// a no-op logger (zerolog.Nop()).
func FromContext(ctx context.Context) *Logger {
	if l, ok := ctx.Value(loggerCtxKey).(*Logger); ok {
		return l
	}
	noop := zerolog.Nop()
	return &Logger{logger: &noop}
}

// Debug logs a message at Debug level.
func (l *Logger) Debug(msg string) {
	l.logger.Debug().Msg(msg)
}

// Debugf logs a formatted message at Debug level.
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.logger.Debug().Msgf(format, args...)
}

// Info logs a message at Info level.
func (l *Logger) Info(msg string) {
	l.logger.Info().Msg(msg)
}

// Infof logs a formatted message at Info level.
func (l *Logger) Infof(format string, args ...interface{}) {
	l.logger.Info().Msgf(format, args...)
}

// Warn logs a message at Warn level.
func (l *Logger) Warn(msg string) {
	l.logger.Warn().Msg(msg)
}

// Warnf logs a formatted message at Warn level.
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.logger.Warn().Msgf(format, args...)
}

// Error logs a message at Error level.
func (l *Logger) Error(msg string) {
	l.logger.Error().Msg(msg)
}

// Errorf logs a formatted message at Error level.
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logger.Error().Msgf(format, args...)
}

// Fatal logs a message at Fatal level and exits.
func (l *Logger) Fatal(msg string) {
	l.logger.Fatal().Msg(msg)
}

// Fatalf logs a formatted message at Fatal level and exits.
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatal().Msgf(format, args...)
}

// Panic logs a message at Panic level and panics.
func (l *Logger) Panic(msg string) {
	l.logger.Panic().Msg(msg)
}

// Panicf logs a formatted message at Panic level and panics.
func (l *Logger) Panicf(format string, args ...interface{}) {
	l.logger.Panic().Msgf(format, args...)
}

// LogHTTPRequest logs an HTTP-style request with common fields:
// method, path, status, duration_ms and client_ip. The level is chosen based
// on the status code (>=500 => error, >=400 => warn, otherwise info).
func (l *Logger) LogHTTPRequest(method, path string, statusCode int, duration time.Duration, clientIP string) {
	var event *zerolog.Event
	if statusCode >= 500 {
		event = l.logger.Error()
	} else if statusCode >= 400 {
		event = l.logger.Warn()
	} else {
		event = l.logger.Info()
	}

	event.
		Str("method", method).
		Str("path", path).
		Int("status", statusCode).
		Dur("duration_ms", duration).
		Str("client_ip", clientIP).
		Msg("http_request")
}

// GetZerolog exposes the underlying zerolog.Logger pointer for advanced use.
func (l *Logger) GetZerolog() *zerolog.Logger {
	return l.logger
}

// SetLevel changes the logger level at runtime. Invalid levels are ignored.
func (l *Logger) SetLevel(level string) {
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		return
	}
	*l.logger = l.logger.Level(lvl)
}

// LogGRPCRequest logs common gRPC request fields. It chooses a level based
// on the presence of an error or non-OK code. Fields include grpc_method,
// duration_ms and grpc_code. If err is non-nil it is attached.
func (l *Logger) LogGRPCRequest(method string, duration time.Duration, code string, err error) {
	var event *zerolog.Event
	if err != nil {
		event = l.logger.Error()
	} else if code != "OK" {
		event = l.logger.Warn()
	} else {
		event = l.logger.Info()
	}

	e := event.
		Str("grpc_method", method).
		Dur("duration_ms", duration).
		Str("grpc_code", code)

	if err != nil {
		e = e.Err(err)
	}

	e.Msg("grpc_request")
}

// WithGRPCMethod returns a new Logger with a grpc_method field attached.
func (l *Logger) WithGRPCMethod(method string) *Logger {
	newLogger := l.logger.With().Str("grpc_method", method).Logger()
	return &Logger{logger: &newLogger}
}

// WithPeerAddr returns a new Logger with a peer_addr field attached.
func (l *Logger) WithPeerAddr(addr string) *Logger {
	newLogger := l.logger.With().Str("peer_addr", addr).Logger()
	return &Logger{logger: &newLogger}
}

// NewDevelopment returns a preconfigured development logger (pretty, debug).
func NewDevelopment(serviceName string) *Logger {
	return New(Config{
		ServiceName: serviceName,
		Environment: EnvironmentDev,
		Level:       "debug",
		Pretty:      true,
	})
}

// NewStage returns a preconfigured staging logger (pretty, info).
func NewStage(serviceName string) *Logger {
	return New(Config{
		ServiceName: serviceName,
		Environment: EnvironmentStage,
		Level:       "info",
		Pretty:      false,
	})
}

// NewProduction returns a preconfigured production logger (json, error).
func NewProduction(serviceName string) *Logger {
	return New(Config{
		ServiceName: serviceName,
		Environment: EnvironmentProd,
		Level:       "warn",
		Pretty:      false,
	})
}

// NewTesting returns a no-op logger suitable for tests.
func NewTesting() *Logger {
	noop := zerolog.Nop()
	return &Logger{logger: &noop}
}

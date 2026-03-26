package logger

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	devEnv  = "dev"
	prodEnv = "prod"
)

type Config struct {
	Env string `env:"LOGGER" env-required:"true"`
}

func New(cfg *Config) (*zap.Logger, error) {
	switch cfg.Env {
	case devEnv:
		loggerConfig := zap.NewDevelopmentConfig()

		loggerConfig.DisableCaller = true
		loggerConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		loggerConfig.EncoderConfig.LineEnding = "\n\n"
		loggerConfig.EncoderConfig.ConsoleSeparator = " | "
		loggerConfig.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString("\033[36m" + t.Format("15:04:05") + "\033[0m")
		}

		logger, err := loggerConfig.Build()
		if err != nil {
			return nil, err
		}

		return logger, nil

	case prodEnv:
		logger, err := zap.NewProduction()
		if err != nil {
			return nil, err
		}

		return logger, nil

	default:
		return nil, fmt.Errorf("unknown environment: %s", cfg.Env)
	}
}

func MiddlewareLogger(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			logger.Info("request started",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
			)

			next.ServeHTTP(ww, r)

			logger.Info("request completed",
				zap.Int("status", ww.Status()),
				zap.Duration("duration", time.Since(start)),
			)
		})
	}
}

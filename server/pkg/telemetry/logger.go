package telemetry

import (
	"context"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"semki/internal/utils/config"
)

var Log *zap.Logger

func SetupLogger(cfg *config.Config) {
	writeSyncer := zapcore.AddSync(os.Stdout)

	encoderConfig := zapcore.EncoderConfig{
		MessageKey:  "message",
		LevelKey:    "level",
		TimeKey:     "ts",
		EncodeTime:  zapcore.ISO8601TimeEncoder,
		EncodeLevel: zapcore.CapitalLevelEncoder,
	}

	var encoder zapcore.Encoder

	if cfg.IsDebug && !cfg.JsonLog {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	core := zapcore.NewCore(encoder, writeSyncer, zap.DebugLevel)
	Log = zap.New(core)
}

//func LoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
//	return func(c *gin.Context) {
//		// Start timer
//		startTime := time.Now()
//
//		// Process request
//		c.Next()
//
//		// End timer
//		endTime := time.Now()
//		latency := endTime.Sub(startTime)
//
//		// Get request details
//		method := c.Request.Method
//		path := c.Request.URL.Path
//		statusCode := c.Writer.Status()
//		clientIP := c.ClientIP()
//
//		// Get trace from context
//		traceID := "No trace ID"
//		span := trace.SpanFromContext(c.Request.Context())
//		if span.SpanContext().IsValid() {
//			traceID = span.SpanContext().TraceID().String()
//		}
//
//		// Log the request details
//		logger.Info("incoming request",
//			zap.String("clientIP", clientIP),
//			zap.String("time", endTime.Format(time.RFC1123)),
//			zap.String("method", method),
//			zap.String("path", path),
//			zap.Int("status", statusCode),
//			zap.Duration("latency", latency),
//			zap.String("traceID", traceID),
//		)
//	}
//}

func TraceForZapLog(ctx context.Context) zap.Field {
	traceID := "No trace ID"
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		traceID = span.SpanContext().TraceID().String()
	}
	return zap.String("traceID", traceID)
}

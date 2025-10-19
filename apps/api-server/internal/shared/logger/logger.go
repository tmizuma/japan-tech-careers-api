package logger

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ContextKey string

const TraceID = ContextKey("trace_id")

var log *zap.Logger

func init() {
	var err error

	conf := zap.Config{
		Level:       zap.NewAtomicLevel(),
		Development: false,
		Encoding:    "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			NameKey:        "name",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	log, err = conf.Build()
	if err != nil {
		panic(err)
	}
}

func Info(ctx context.Context, message string, fields ...zap.Field) {
	fields = append(fields, zap.String("trace_id", retrieveTraceID(ctx)))
	log.Info(message, fields...)
}

func Debug(ctx context.Context, message string, fields ...zap.Field) {
	fields = append(fields, zap.String("trace_id", retrieveTraceID(ctx)))
	log.Debug(message, fields...)
}

func Error(ctx context.Context, message string, fields ...zap.Field) {
	fields = append(fields, zap.String("trace_id", retrieveTraceID(ctx)))
	log.Error(message, fields...)
}

func Warn(ctx context.Context, message string, fields ...zap.Field) {
	fields = append(fields, zap.String("trace_id", retrieveTraceID(ctx)))
	log.Warn(message, fields...)
}

func retrieveTraceID(ctx context.Context) string {
	traceID, _ := ctx.Value(TraceID).(string)
	return traceID
}

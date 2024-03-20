package log

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.elastic.co/apm"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger LoggerConf

type LoggerConf struct {
	// Logger instance
	dep *zap.Logger
}

type Logger interface {
	Info(ctx context.Context, msg string, meta interface{})
	Error(ctx context.Context, msg string, meta interface{})
}

func SetupLogger() *LoggerConf {
	lg := zap.NewProductionConfig()
	lg.DisableStacktrace = false
	lg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		MessageKey:   "msg",
		LevelKey:     "level",
		TimeKey:      "ts",
		CallerKey:    "caller",
		EncodeCaller: zapcore.ShortCallerEncoder,
		EncodeLevel:  zapcore.CapitalColorLevelEncoder,
		EncodeTime:   zapcore.ISO8601TimeEncoder,
		EncodeDuration: func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendInt64(int64(d) / int64(time.Millisecond))
		},
		EncodeName: zapcore.FullNameEncoder,
	})

	lg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	lg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	lg.EncoderConfig.EncodeDuration = func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendInt64(int64(d) / int64(time.Millisecond))
	}
	lg.EncoderConfig.EncodeName = zapcore.FullNameEncoder

	core := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zap.DebugLevel)

	zapLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	return &LoggerConf{
		dep: zapLogger,
	}
}

func Init(l *LoggerConf) {
	logger = LoggerConf{
		dep: l.dep,
	}
}

func GetLogger() Logger {
	return &logger
}

func (l *LoggerConf) Info(ctx context.Context, msg string, meta interface{}) {
	metaField := zap.Any("meta", meta)
	l.withTraceInfo(ctx).dep.Info(msg, metaField)
}

func (l *LoggerConf) Error(ctx context.Context, msg string, meta interface{}) {
	metaField := zap.Any("meta", meta)
	l.withTraceInfo(ctx).dep.Error(msg, metaField)
}

func (l *LoggerConf) withTraceInfo(ctx context.Context) *LoggerConf {
	span := apm.SpanFromContext(ctx)
	if span == nil {
		return l.Clone(l.dep)
	}
	apmCtx := span.TraceContext()
	traceId := ZapString("trace.id", fmt.Sprintf("%s", apmCtx.Trace.String()))
	spanId := ZapString("span.id", fmt.Sprintf("%s", apmCtx.Span.String()))
	return l.Clone(l.dep.With(
		traceId,
		spanId,
	))
}

func ZapString(key, value string) zap.Field {
	if value == "" {
		return zap.Skip()
	}
	return zap.String(key, value)
}

// Clone will create new Logger instance with specified zap.Logger.
func (l *LoggerConf) Clone(logger *zap.Logger) *LoggerConf {
	return &LoggerConf{
		dep: logger,
	}
}

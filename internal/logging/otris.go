package logging

import (
	"bytes"
	"context"
	"github.com/Totus-Floreo/otris"
	"github.com/jackc/pgx/v5/tracelog"
	"log/slog"
	"os"
)

const layout = "15:04"

type Logger struct {
	*slog.Logger
}

func NewLogger() Logger {
	handler := otris.NewHandlerBuilder().WithPretty().WithWriter(os.Stdout).Build()
	logger := slog.New(handler)
	slog.SetDefault(logger)
	return Logger{logger}
}

func (l Logger) Write(p []byte) (n int, err error) {
	buffer := bytes.NewBuffer(p)
	l.Logger.Info("Event", buffer.String())
	return len(p), nil
}

type DBlogger struct {
	Logger
}

func (l *DBlogger) Log(_ context.Context, level tracelog.LogLevel, msg string, data map[string]any) {
	fields := make([]any, len(data))
	i := 0
	for k, v := range data {
		fields[i] = slog.Any(k, v)
		i++
	}

	switch level {
	case tracelog.LogLevelTrace:
		l.Logger.Debug(msg, append(fields, slog.String("PGX_LOG_LEVEL", level.String()))...)
	case tracelog.LogLevelDebug:
		l.Logger.Debug(msg, fields...)
	case tracelog.LogLevelInfo:
		l.Logger.Info(msg, fields...)
	case tracelog.LogLevelWarn:
		l.Logger.Warn(msg, fields...)
	case tracelog.LogLevelError:
		l.Logger.Error(msg, fields...)
	default:
		l.Logger.Error(msg, append(fields, slog.String("PGX_LOG_LEVEL", level.String()))...)
	}
}

func NewTraceLog(log Logger) *tracelog.TraceLog {
	return &tracelog.TraceLog{
		Logger:   &DBlogger{log},
		LogLevel: tracelog.LogLevelTrace,
	}
}

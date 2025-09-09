package logger

import (
	"os"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
	Fatal(msg string)

	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
	WithError(err error) Logger
}

type logrusLogger struct {
	logger *logrus.Logger
	entry  *logrus.Entry
}

type zapLogger struct {
	logger *zap.Logger
	sugar  *zap.SugaredLogger
}

var (
	globalLogger Logger
	once         sync.Once
)

// InitLogger initializes the global logger
func InitLogger(logLevel string, format string) error {
	var err error
	once.Do(func() {
		switch strings.ToLower(format) {
		case "zap":
			globalLogger, err = newZapLogger(logLevel)
		default:
			globalLogger, err = newLogrusLogger(logLevel, format)
		}
	})
	return err
}

// GetLogger returns the global logger instance
func GetLogger() Logger {
	if globalLogger == nil {
		// Default to logrus if not initialized
		globalLogger, _ = newLogrusLogger("info", "json")
	}
	return globalLogger
}

// Logrus implementation
func newLogrusLogger(logLevel, format string) (Logger, error) {
	logger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Set formatter
	switch strings.ToLower(format) {
	case "text":
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
			ForceColors:   true,
		})
	default:
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z",
		})
	}

	// Set output
	logger.SetOutput(os.Stdout)

	return &logrusLogger{
		logger: logger,
		entry:  logrus.NewEntry(logger),
	}, nil
}

func (l *logrusLogger) Debug(msg string) {
	l.entry.Debug(msg)
}

func (l *logrusLogger) Info(msg string) {
	l.entry.Info(msg)
}

func (l *logrusLogger) Warn(msg string) {
	l.entry.Warn(msg)
}

func (l *logrusLogger) Error(msg string) {
	l.entry.Error(msg)
}

func (l *logrusLogger) Fatal(msg string) {
	l.entry.Fatal(msg)
}

func (l *logrusLogger) WithField(key string, value interface{}) Logger {
	return &logrusLogger{
		logger: l.logger,
		entry:  l.entry.WithField(key, value),
	}
}

func (l *logrusLogger) WithFields(fields map[string]interface{}) Logger {
	return &logrusLogger{
		logger: l.logger,
		entry:  l.entry.WithFields(fields),
	}
}

func (l *logrusLogger) WithError(err error) Logger {
	return &logrusLogger{
		logger: l.logger,
		entry:  l.entry.WithError(err),
	}
}

// Zap implementation
func newZapLogger(logLevel string) (Logger, error) {
	level := zap.InfoLevel
	if parsedLevel, err := zapcore.ParseLevel(logLevel); err == nil {
		level = parsedLevel
	}

	config := zap.Config{
		Level:       zap.NewAtomicLevelAt(level),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return &zapLogger{
		logger: logger,
		sugar:  logger.Sugar(),
	}, nil
}

func (l *zapLogger) Debug(msg string) {
	l.sugar.Debug(msg)
}

func (l *zapLogger) Info(msg string) {
	l.sugar.Info(msg)
}

func (l *zapLogger) Warn(msg string) {
	l.sugar.Warn(msg)
}

func (l *zapLogger) Error(msg string) {
	l.sugar.Error(msg)
}

func (l *zapLogger) Fatal(msg string) {
	l.sugar.Fatal(msg)
}

func (l *zapLogger) WithField(key string, value interface{}) Logger {
	return &zapLogger{
		logger: l.logger.With(zap.Any(key, value)),
		sugar:  l.logger.With(zap.Any(key, value)).Sugar(),
	}
}

func (l *zapLogger) WithFields(fields map[string]interface{}) Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}

	return &zapLogger{
		logger: l.logger.With(zapFields...),
		sugar:  l.logger.With(zapFields...).Sugar(),
	}
}

func (l *zapLogger) WithError(err error) Logger {
	return &zapLogger{
		logger: l.logger.With(zap.Error(err)),
		sugar:  l.logger.With(zap.Error(err)).Sugar(),
	}
}

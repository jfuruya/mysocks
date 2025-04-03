package mysocks

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func init() {
	logger, _ = zap.NewDevelopment()
}

// ログレベルを定義
const (
	logLevelDebug = iota
	logLevelInfo
	logLevelWarn
	logLevelError
	logLevelFatal
)

func logDebug(message string, fields map[string]interface{}) {
	logWithLevel(logLevelDebug, message, fields)
}

func logInfo(message string, fields map[string]interface{}) {
	logWithLevel(logLevelInfo, message, fields)
}

func logWarn(message string, fields map[string]interface{}) {
	logWithLevel(logLevelWarn, message, fields)
}

func logError(message string, fields map[string]interface{}) {
	logWithLevel(logLevelError, message, fields)
}

func logFatal(message string, fields map[string]interface{}) {
	logWithLevel(logLevelFatal, message, fields)
}

func logWithLevel(level int, message string, fields map[string]interface{}) {
	var zapLevel zapcore.Level
	switch level {
	case logLevelDebug:
		zapLevel = zap.DebugLevel
	case logLevelInfo:
		zapLevel = zap.InfoLevel
	case logLevelWarn:
		zapLevel = zap.WarnLevel
	case logLevelError:
		zapLevel = zap.ErrorLevel
	case logLevelFatal:
		zapLevel = zap.FatalLevel
	}
	zapFields := toZapFields(fields)
	logger.Log(zapLevel, message, zapFields...)
}

func toZapFields(fields map[string]interface{}) []zap.Field {
	var zapFields []zap.Field
	for key, value := range fields {
		zapFields = append(zapFields, zap.Any(key, value))
	}
	return zapFields
}

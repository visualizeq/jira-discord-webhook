package utils

import (
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewZapLoggerWithRotate returns a zap.Logger and the rotatelogs writer for reuse.
func NewZapLoggerWithRotate(logPath string, level zap.AtomicLevel) (*zap.Logger, *rotatelogs.RotateLogs, error) {
	writer, err := rotatelogs.New(
		logPath+".%Y-%m-%d",
		rotatelogs.WithLinkName(logPath),
		rotatelogs.WithMaxAge(7*24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour),
	)
	if err != nil {
		return nil, nil, err
	}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(writer),
		level,
	)
	return zap.New(core), writer, nil
}

// NewRotateWriter returns a rotatelogs writer with the same rotation config as zap logger.
func NewRotateWriter(logPath string) (*rotatelogs.RotateLogs, error) {
	return rotatelogs.New(
		logPath+".%Y-%m-%d",
		rotatelogs.WithLinkName(logPath),
		rotatelogs.WithMaxAge(7*24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour),
	)
}

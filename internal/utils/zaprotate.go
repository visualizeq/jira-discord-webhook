package utils

import (
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewZapLoggerWithRotate returns a zap.Logger that writes to a file with daily rotation and 7 days retention.
func NewZapLoggerWithRotate(logPath string, level zap.AtomicLevel) (*zap.Logger, error) {
	writer, err := rotatelogs.New(
		logPath+".%Y-%m-%d",
		rotatelogs.WithLinkName(logPath),
		rotatelogs.WithMaxAge(7*24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour),
	)
	if err != nil {
		return nil, err
	}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(writer),
		level,
	)
	return zap.New(core), nil
}

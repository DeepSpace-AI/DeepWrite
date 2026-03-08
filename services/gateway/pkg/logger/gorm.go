package logger

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type GormLogger struct {
	SlowThreshold time.Duration
	LogLevel      gormlogger.LogLevel
}

func NewGormLogger(level gormlogger.LogLevel) gormlogger.Interface {
	return &GormLogger{
		SlowThreshold: 200 * time.Millisecond,
		LogLevel:      level,
	}
}

func (gl *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := *gl
	newLogger.LogLevel = level
	return &newLogger
}

func (gl *GormLogger) Info(_ context.Context, msg string, args ...any) {
	if gl.LogLevel < gormlogger.Info {
		return
	}
	S().Infof(msg, args...)
}

func (gl *GormLogger) Warn(_ context.Context, msg string, args ...any) {
	if gl.LogLevel < gormlogger.Warn {
		return
	}
	S().Warnf(msg, args...)
}

func (gl *GormLogger) Error(_ context.Context, msg string, args ...any) {
	if gl.LogLevel < gormlogger.Error {
		return
	}
	S().Errorf(msg, args...)
}

func (gl *GormLogger) Trace(_ context.Context, begin time.Time, fc func() (string, int64), err error) {
	if gl.LogLevel == gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	switch {
	case err != nil && gl.LogLevel >= gormlogger.Error && !errors.Is(err, gorm.ErrRecordNotFound):
		L().Error("gorm query",
			zap.Error(err),
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		)
	case gl.SlowThreshold > 0 && elapsed > gl.SlowThreshold && gl.LogLevel >= gormlogger.Warn:
		L().Warn("gorm slow query",
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		)
	case gl.LogLevel >= gormlogger.Info:
		L().Info("gorm query",
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		)
	}
}

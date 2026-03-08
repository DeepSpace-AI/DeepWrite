package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Config struct {
	Dir        string
	Filename   string
	Level      string
	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int
	Compress   bool
}

var (
	defaultLogger *zap.Logger
	defaultSugar  *zap.SugaredLogger
	initOnce      sync.Once
)

func defaultConfig() Config {
	return Config{
		Dir:        "logs",
		Filename:   "gateway",
		Level:      "info",
		MaxSizeMB:  100,
		MaxBackups: 30,
		MaxAgeDays: 30,
		Compress:   true,
	}
}

func InitLogger(appEnv string, customCfg ...Config) error {
	var initErr error

	initOnce.Do(func() {
		cfg := defaultConfig()
		if len(customCfg) > 0 {
			cfg = customCfg[0]
		}

		if cfg.Dir == "" {
			cfg.Dir = "logs"
		}
		if cfg.Filename == "" {
			cfg.Filename = "gateway"
		}
		if cfg.MaxSizeMB <= 0 {
			cfg.MaxSizeMB = 100
		}
		if cfg.MaxBackups <= 0 {
			cfg.MaxBackups = 30
		}
		if cfg.MaxAgeDays <= 0 {
			cfg.MaxAgeDays = 30
		}
		if appEnv == "development" {
			if cfg.Level == "" || cfg.Level == "info" {
				cfg.Level = "debug"
			}
		}
		if cfg.Level == "" {
			cfg.Level = "info"
		}

		if err := os.MkdirAll(cfg.Dir, 0o755); err != nil {
			initErr = fmt.Errorf("create log dir failed: %w", err)
			return
		}

		writer := newDailySizeWriter(cfg)
		level := parseLevel(cfg.Level)

		encoderCfg := zap.NewProductionEncoderConfig()
		encoderCfg.TimeKey = "time"
		encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderCfg.EncodeLevel = zapcore.LowercaseLevelEncoder

		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.NewMultiWriteSyncer(
				zapcore.AddSync(os.Stdout),
				zapcore.AddSync(writer),
			),
			level,
		)

		defaultLogger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
		defaultSugar = defaultLogger.Sugar()
	})

	return initErr
}

func L() *zap.Logger {
	if defaultLogger == nil {
		_ = InitLogger("development")
	}
	return defaultLogger
}

func S() *zap.SugaredLogger {
	if defaultSugar == nil {
		_ = InitLogger("development")
	}
	return defaultSugar
}

func Sync() {
	if defaultLogger != nil {
		_ = defaultLogger.Sync()
	}
}

func parseLevel(level string) zapcore.LevelEnabler {
	switch level {
	case "debug":
		return zap.DebugLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	default:
		return zap.InfoLevel
	}
}

type dailySizeWriter struct {
	cfg         Config
	mu          sync.Mutex
	currentDate string
	current     *lumberjack.Logger
}

func newDailySizeWriter(cfg Config) *dailySizeWriter {
	w := &dailySizeWriter{cfg: cfg}
	w.rotateIfNeeded(time.Now())
	return w
}

func (w *dailySizeWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.rotateIfNeeded(time.Now())
	return w.current.Write(p)
}

func (w *dailySizeWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.current != nil {
		return w.current.Close()
	}
	return nil
}

func (w *dailySizeWriter) rotateIfNeeded(now time.Time) {
	date := now.Format("2006-01-02")
	if date == w.currentDate && w.current != nil {
		return
	}

	if w.current != nil {
		_ = w.current.Close()
	}

	filename := filepath.Join(w.cfg.Dir, fmt.Sprintf("%s-%s.log", w.cfg.Filename, date))
	w.currentDate = date
	w.current = &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    w.cfg.MaxSizeMB,
		MaxBackups: w.cfg.MaxBackups,
		MaxAge:     w.cfg.MaxAgeDays,
		Compress:   w.cfg.Compress,
		LocalTime:  true,
	}
}

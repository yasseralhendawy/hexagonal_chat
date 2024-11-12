package zaplogger

import (
	"fmt"
	"os"
	"time"

	"github.com/yasseralhendawy/hexagonal_chat/config"
	logger "github.com/yasseralhendawy/hexagonal_chat/pkg/logger/adapter"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type ZapLog struct {
	logger *zap.SugaredLogger
}

func CreateLogger(cfg *config.Logger) logger.Logger {
	var z ZapLog
	z.Init(cfg)
	return &z
}

func (z *ZapLog) Init(cfg *config.Logger) {
	stdout := zapcore.AddSync(os.Stdout)
	fileName := fmt.Sprintf("%s-%s.%s", cfg.FilePath, time.Now().Format("2006-01-02"), "log")
	file := zapcore.AddSync(&lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    10, // megabytes
		MaxBackups: 3,
		MaxAge:     7, // days
		LocalTime:  true,
	})

	level := zap.NewAtomicLevelAt(getLevel(cfg))

	productionCfg := zap.NewProductionEncoderConfig()
	productionCfg.TimeKey = "timestamp"
	productionCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	developmentCfg := zap.NewDevelopmentEncoderConfig()
	developmentCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	consoleEncoder := zapcore.NewConsoleEncoder(developmentCfg)
	fileEncoder := zapcore.NewJSONEncoder(productionCfg)

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, stdout, level),
		zapcore.NewCore(fileEncoder, file, level),
	)

	z.logger = zap.New(core).Sugar()

	defer z.logger.Sync()
}

// Debug implements Logger.
func (z *ZapLog) Debug(lf logger.LoggerFrom, msg string, extra map[string]interface{}) {
	args := prepareArgs(lf, extra)
	z.logger.Debugw(msg, args...)
}

// Error implements Logger.
func (z *ZapLog) Error(lf logger.LoggerFrom, msg string, extra map[string]interface{}) {
	args := prepareArgs(lf, extra)
	z.logger.Errorw(msg, args...)
}

// Fatal implements Logger.
func (z *ZapLog) Fatal(lf logger.LoggerFrom, msg string, extra map[string]interface{}) {
	args := prepareArgs(lf, extra)
	z.logger.Fatalw(msg, args...)
}

// Warn implements Logger.
func (z *ZapLog) Warn(lf logger.LoggerFrom, msg string, extra map[string]interface{}) {
	args := prepareArgs(lf, extra)
	z.logger.Warnw(msg, args...)
}

// Info implements Logger.
func (z *ZapLog) Info(lf logger.LoggerFrom, msg string, extra map[string]interface{}) {
	args := prepareArgs(lf, extra)
	z.logger.Infow(msg, args...)
}

func getLevel(cfg *config.Logger) zapcore.Level {
	switch cfg.Level {
	case -1:
		return zap.DebugLevel
	case 0:
		return zapcore.InfoLevel
	case 1:
		return zapcore.WarnLevel
	case 2:
		return zapcore.ErrorLevel
	case 3:
		return zapcore.DPanicLevel
	case 4:
		return zapcore.PanicLevel
	case 5:
		return zapcore.FatalLevel
	default:
		return zapcore.FatalLevel
	}
}

func prepareArgs(lf logger.LoggerFrom, extra map[string]interface{}) []interface{} {
	if extra == nil {
		extra = make(map[string]interface{})
	}
	extra["log-from"] = lf

	params := make([]interface{}, 0, len(extra))

	for k, v := range extra {
		params = append(params, k)
		params = append(params, v)
	}

	return params
}

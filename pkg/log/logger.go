package log

import (
	"fmt"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"os"
	"path"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	once            sync.Once
	logger          *Logger
	logLevel        = zap.DebugLevel // 默认日志打印级别
	defaultLogsPath = "logs"         // 默认日志保存文件夹
)

// NewLogger .
func NewLogger() *Logger {
	once.Do(func() {
		logger = new(Logger)
		logger.base = zap.New(newZapHook(), zap.AddCaller()).Sugar() // 开启日志中进行代码文件追踪
	})
	return logger
}

// newEncoderConfig .
func newEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		// Keys can be anything except the empty string.
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

// TimeEncoder .
func TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("[2006-01-02 15:04:05]"))
}

// newZapHook create hook
func newZapHook() zapcore.Core {
	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewConsoleEncoder(newEncoderConfig()), zapcore.AddSync(getFileWriter()), getLevel()),    // 打印在文件中
		zapcore.NewCore(zapcore.NewConsoleEncoder(newEncoderConfig()), zapcore.AddSync(getConsoleWriter()), getLevel()), // 打印在控制台
	)
	return core
}

// getLevel .
func getLevel() zap.LevelEnablerFunc {
	currentLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= logLevel
	})
	return currentLevel
}

// getConsoleWriter .
func getConsoleWriter() zapcore.WriteSyncer {
	return zapcore.Lock(os.Stdout)
}

// getFileWriter .
func getFileWriter() *rotatelogs.RotateLogs {
	baseLogPath := path.Join(defaultLogsPath)
	_, err := os.Stat(defaultLogsPath)
	if err != nil {
		if os.IsNotExist(err) {
			_ = os.MkdirAll(baseLogPath, os.ModePerm)
		}
	}
	writer, err := rotatelogs.New(
		path.Join(baseLogPath, "%Y-%m-%d-%H.log"),
		rotatelogs.WithMaxAge(30*24*time.Hour),
		rotatelogs.WithRotationTime(1*time.Hour),
	)
	if err != nil {
		panic(err)
	}
	return writer
}

// Logger .
type Logger struct {
	base *zap.SugaredLogger
}

// Debug uses fmt.Sprint to construct and log a message.
func (l *Logger) Debug(args ...interface{}) {
	l.base.Debug(args...)
}

// Info uses fmt.Sprint to construct and log a message.
func (l *Logger) Info(args ...interface{}) {
	l.base.Info(args...)
}

// Warn uses fmt.Sprint to construct and log a message.
func (l *Logger) Warn(args ...interface{}) {
	l.base.Warn(args...)
}

// Error uses fmt.Sprint to construct and log a message.
func (l *Logger) Error(args ...interface{}) {
	l.base.Error(args...)
}

// DPanic uses fmt.Sprint to construct and log a message. In development, the
// logger then panics. (See DPanicLevel for details.)
func (l *Logger) DPanic(args ...interface{}) {
	l.base.DPanic(args...)
}

// Panic uses fmt.Sprint to construct and log a message, then panics.
func (l *Logger) Panic(args ...interface{}) {
	l.base.Panic(args...)
}

// Fatal uses fmt.Sprint to construct and log a message, then calls os.Exit.
func (l *Logger) Fatal(args ...interface{}) {
	l.base.Fatal(args...)
}

// Printf must have the same semantics as log.Printf.
func (l *Logger) Printf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

// Debugf uses fmt.Sprintf to log a templated message.
func (l *Logger) Debugf(template string, args ...interface{}) {
	l.base.Debugf(template, args...)
}

// Infof uses fmt.Sprintf to log a templated message.
func (l *Logger) Infof(template string, args ...interface{}) {
	l.base.Infof(template, args...)
}

// Warnf uses fmt.Sprintf to log a templated message.
func (l *Logger) Warnf(template string, args ...interface{}) {
	l.base.Warnf(template, args...)
}

// Errorf uses fmt.Sprintf to log a templated message.
func (l *Logger) Errorf(template string, args ...interface{}) {
	l.base.Errorf(template, args...)
}

// DPanicf uses fmt.Sprintf to log a templated message. In development, the
// logger then panics. (See DPanicLevel for details.)
func (l *Logger) DPanicf(template string, args ...interface{}) {
	l.base.DPanicf(template, args...)
}

// Panicf uses fmt.Sprintf to log a templated message, then panics.
func (l *Logger) Panicf(template string, args ...interface{}) {
	l.base.Panicf(template, args...)
}

// Fatalf uses fmt.Sprintf to log a templated message, then calls os.Exit.
func (l *Logger) Fatalf(template string, args ...interface{}) {
	l.base.Fatalf(template, args...)
}

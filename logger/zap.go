package logger

import (
	"context"
	"fmt"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
	"time"
)

type config struct {
	level         zapcore.Level
	syncWriter    io.Writer
	debugMode     bool
	rotationTime  time.Duration
	logExpireTime time.Duration
}

// ConfigOption define callback for setup
type ConfigOption func(*config)

// WithLevel setup log level
func WithLevel(level zapcore.Level) ConfigOption {
	return func(c *config) {
		c.level = level
	}
}

func WithRunMode(debugMode bool) ConfigOption {
	return func(c *config) {
		c.debugMode = debugMode
	}
}

func WithSyncWriter(syncWriter io.Writer) ConfigOption {
	return func(c *config) {
		c.syncWriter = syncWriter
	}
}

//日志切割时间
func WithRotationTime(t time.Duration) ConfigOption {
	return func(c *config) {
		c.logExpireTime = t
	}
}

//日志最大有效时间
func WithLogExpire(t time.Duration) ConfigOption {
	return func(c *config) {
		c.rotationTime = t
	}
}

func getDefaultConfig() zapcore.EncoderConfig {
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder
	return cfg
}

func getDefaultWriter(filename string, expireTime time.Duration, rotationTime time.Duration) io.Writer {
	hook, err := rotatelogs.New(
		filename,
		rotatelogs.WithMaxAge(expireTime),
		rotatelogs.WithRotationTime(rotationTime),
	)

	if err != nil {
		panic(err)
	}

	return hook
}

func NewLogger(serviceName, filePath, fileName string, opts ...ConfigOption) *zap.Logger {
	fileFullName := filePath + fileName
	defaultCfg := config{
		level:         zapcore.DebugLevel,
		debugMode:     false,
		rotationTime:  time.Hour,
		logExpireTime: 24 * time.Hour,
	}

	for _, o := range opts {
		o(&defaultCfg)
	}

	defaultCfg.syncWriter = getDefaultWriter(fileFullName, defaultCfg.logExpireTime, defaultCfg.rotationTime)

	encoder := getDefaultConfig()

	core := zapcore.NewTee(
		zapcore.NewCore( //文件，json
			zapcore.NewJSONEncoder(encoder),
			zapcore.AddSync(defaultCfg.syncWriter),
			zap.NewAtomicLevelAt(defaultCfg.level),
		),
		zapcore.NewCore( // 控制台
			zapcore.NewConsoleEncoder(encoder),
			zapcore.AddSync(os.Stdout),
			zap.NewAtomicLevelAt(defaultCfg.level),
		),
	)

	var logger *zap.Logger

	filed := zap.Fields(zap.String("service_name", serviceName))
	if defaultCfg.debugMode == true {
		logger = zap.New(core,
			// 开启开发模式，堆栈跟踪
			zap.AddCaller(), zap.AddCallerSkip(2),
			//  开启文件及行号
			zap.Development())
	} else {
		logger = zap.New(core, filed)
	}

	return logger
}

var DefaultMessageKey = "msg"

var _ ILogger = (*ZLogger)(nil)

func NewWitchContextZLogger(ctx context.Context, z *zap.Logger, kvs ...interface{}) ILogger {
	return &ZLogger{zLog: z, msgKey: DefaultMessageKey, kvs: kvs, ctx: ctx}
}

func NewZLogger(z *zap.Logger, kvs ...interface{}) ILogger {
	return &ZLogger{zLog: z, msgKey: DefaultMessageKey, kvs: kvs}
}

type ZLogger struct {
	zLog   *zap.Logger
	msgKey string
	ctx    context.Context
	kvs    []interface{}
}

func (z *ZLogger) Debugw(ctx context.Context, format string, a ...interface{}) {
	z.WithContext(ctx).Debugf(format, a...)
}

func (z *ZLogger) Infow(ctx context.Context, format string, a ...interface{}) {
	z.WithContext(ctx).Infof(format, a...)
}

func (z *ZLogger) Warnw(ctx context.Context, format string, a ...interface{}) {
	z.WithContext(ctx).Warnf(format, a...)
}

func (z *ZLogger) Errorw(ctx context.Context, format string, a ...interface{}) {
	z.WithContext(ctx).Errorf(format, a...)
}

func (z *ZLogger) WithKeyValue(kvs ...interface{}) ILogger {
	z.kvs = append(z.kvs, kvs...)
	return z
}

func (z *ZLogger) WithContext(ctx context.Context) ILogger {
	z.ctx = ctx
	return z
}

func (z *ZLogger) getZapField(keyvals ...interface{}) []zap.Field {
	var data []zap.Field
	keyvals = append(z.kvs, keyvals...)
	bindValues(z.ctx, keyvals)
	for i := 0; i < len(keyvals); i += 2 {
		data = append(data, zap.Any(fmt.Sprint(keyvals[i]), fmt.Sprint(keyvals[i+1])))
	}
	return data
}

func (z *ZLogger) Debug(a ...interface{}) {
	field := z.getZapField(z.msgKey, fmt.Sprint(a...))
	z.zLog.Debug("", field...)
}

func (z *ZLogger) Debugf(format string, a ...interface{}) {
	field := z.getZapField(z.msgKey, fmt.Sprintf(format, a...))
	z.zLog.Debug("", field...)
}

func (z *ZLogger) Info(a ...interface{}) {
	field := z.getZapField(z.msgKey, fmt.Sprint(a...))
	z.zLog.Info("", field...)
}

func (z *ZLogger) Infof(format string, a ...interface{}) {
	field := z.getZapField(z.msgKey, fmt.Sprintf(format, a...))
	z.zLog.Info("", field...)
}

func (z *ZLogger) Warn(a ...interface{}) {
	field := z.getZapField(z.msgKey, fmt.Sprint(a...))
	z.zLog.Warn("", field...)
}

func (z *ZLogger) Warnf(format string, a ...interface{}) {
	field := z.getZapField(z.msgKey, fmt.Sprintf(format, a...))
	z.zLog.Warn("", field...)
}

func (z *ZLogger) Error(a ...interface{}) {
	field := z.getZapField(z.msgKey, fmt.Sprint(a...))
	z.zLog.Error("", field...)
}

func (z *ZLogger) Errorf(format string, a ...interface{}) {
	field := z.getZapField(z.msgKey, fmt.Sprintf(format, a...))
	z.zLog.Error("", field...)
}

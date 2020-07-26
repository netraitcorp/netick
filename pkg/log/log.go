package log

import (
	"fmt"
	"os"
	"sync"

	"github.com/netraitcorp/netick/pkg/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Options struct {
	Env        types.Environment
	Filename   string
	MaxSize    int
	MaxAge     int
	MaxBackups int
	Compress   bool
	LocalTime  bool
	Level      string
}

var (
	once       sync.Once
	logger     *zap.SugaredLogger
	loggerOpts *Options
)

var level = map[string]zapcore.Level{
	"debug":  zapcore.DebugLevel,
	"info":   zapcore.InfoLevel,
	"warn":   zapcore.WarnLevel,
	"error":  zapcore.ErrorLevel,
	"dpanic": zapcore.DPanicLevel,
	"panic":  zapcore.PanicLevel,
	"fatal":  zapcore.FatalLevel,
}

func NewOptions() *Options {
	return &Options{
		Env:        types.EnvProd,
		Filename:   "./logs/netick.log",
		MaxSize:    1024,
		MaxAge:     30,
		MaxBackups: 14,
		Compress:   true,
		LocalTime:  true,
		Level:      "info",
	}
}

func InitLogger(opts *Options) {
	once.Do(func() {
		hook := lumberjack.Logger{
			Filename:   opts.Filename,
			MaxSize:    opts.MaxSize,
			MaxBackups: opts.MaxBackups,
			MaxAge:     opts.MaxAge,
			Compress:   opts.Compress,
		}

		lvl := zap.InfoLevel
		if l, ok := level[opts.Level]; ok {
			lvl = l
		}
		atomicLevel := zap.NewAtomicLevel()
		atomicLevel.SetLevel(lvl)
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

		var mws []zapcore.WriteSyncer
		mws = append(mws, zapcore.AddSync(&hook))
		if opts.Env == types.EnvDev {
			mws = append(mws, zapcore.AddSync(os.Stdout))
		}

		core := zapcore.NewCore(
			//zapcore.NewJSONEncoder(encoderConfig),
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.NewMultiWriteSyncer(mws...),
			atomicLevel,
		)

		logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)).Sugar()
		loggerOpts = opts

		if opts.Env == types.EnvDev {
			fmt.Println("")
			fmt.Println("==> Log data will now stream in as it occurs (in development mode):")
			fmt.Println("")
		}
	})
}

func Debug(template string, args ...interface{}) {
	logger.Debugf(template, args...)
}

func Info(template string, args ...interface{}) {
	logger.Infof(template, args...)
}

func Warn(template string, args ...interface{}) {
	logger.Warnf(template, args...)
}

func Error(template string, args ...interface{}) {
	logger.Errorf(template, args...)
}

func DPanic(template string, args ...interface{}) {
	logger.DPanicf(template, args...)
}

func Panic(template string, args ...interface{}) {
	logger.Panicf(template, args...)
}

func Fatal(template string, args ...interface{}) {
	logger.Fatalf(template, args...)
}

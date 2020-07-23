package log

import (
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Options struct {
	Filename   string
	MaxSize    int
	MaxAge     int
	MaxBackups int
	Compress   bool
	LocalTime  bool
	Level      string
}

var (
	once   sync.Once
	logger *zap.Logger
)

var level = map[string]zapcore.Level{
	"debug":  zap.DebugLevel,
	"info":   zap.InfoLevel,
	"warn":   zap.WarnLevel,
	"error":  zap.ErrorLevel,
	"dpanic": zap.DPanicLevel,
	"panic":  zap.PanicLevel,
	"fatal":  zap.FatalLevel,
}

func NewOptions() *Options {
	return &Options{
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

		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(&hook)),
			atomicLevel,
		)

		logger = zap.New(core, zap.AddCaller(), zap.Development(), zap.Fields(zap.String("serviceName", "serviceName")))
	})
}

func Info() {
	logger.Info("无法获取网址",
		zap.String("url", "http://www.baidu.com"),
		zap.Int("attempt", 3),
		zap.Duration("backoff", time.Second))
}

func Fatal() {

}

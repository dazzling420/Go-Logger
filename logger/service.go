package logger

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/dazzling420/go-logger/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

const (
	INFO   = zap.InfoLevel   // 0
	WARN   = zap.WarnLevel   // 1
	ERROR  = zap.ErrorLevel  // 2
	DPANIC = zap.DPanicLevel // 3
	PANIC  = zap.PanicLevel  // 4
	FATAL  = zap.FatalLevel  // 5
	DEBUG  = zap.DebugLevel  // -1
)

func GetLevel(l string) zapcore.Level {
	switch l {
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	case "DPANIC":
		return PANIC
	case "PANIC":
		return PANIC
	case "FATAL":
		return FATAL
	case "DEBUG":
		return DEBUG
	default:
		return INFO
	}
}

type Field = zap.Field

var (
	Skip        = zap.Skip
	Binary      = zap.Binary
	Bool        = zap.Bool
	Boolp       = zap.Boolp
	ByteString  = zap.ByteString
	Complex128  = zap.Complex128
	Complex128p = zap.Complex128p
	Complex64   = zap.Complex64
	Complex64p  = zap.Complex64p
	Float64     = zap.Float64
	Float64p    = zap.Float64p
	Float32     = zap.Float32
	Float32p    = zap.Float32p
	Int         = zap.Int
	Intp        = zap.Intp
	Int64       = zap.Int64
	Int64p      = zap.Int64p
	Int32       = zap.Int32
	Int32p      = zap.Int32p
	Int16       = zap.Int16
	Int16p      = zap.Int16p
	Int8        = zap.Int8
	Int8p       = zap.Int8p
	String      = zap.String
	Stringp     = zap.Stringp
	Uint        = zap.Uint
	Uintp       = zap.Uintp
	Uint64      = zap.Uint64
	Uint64p     = zap.Uint64p
	Uint32      = zap.Uint32
	Uint32p     = zap.Uint32p
	Uint16      = zap.Uint16
	Uint16p     = zap.Uint16p
	Uint8       = zap.Uint8
	Uint8p      = zap.Uint8p
	Uintptr     = zap.Uintptr
	Uintptrp    = zap.Uintptrp
	Reflect     = zap.Reflect
	Namespace   = zap.Namespace
	Stringer    = zap.Stringer
	Time        = zap.Time
	Timep       = zap.Timep
	Stack       = zap.Stack
	StackSkip   = zap.StackSkip
	Duration    = zap.Duration
	Durationp   = zap.Durationp
	Any         = zap.Any
)

type Service interface {
	GetLogger() *standardLogger
	GetSDLogger() *zap.SugaredLogger
	GetZapLogger() *zap.Logger

	Errorf(format string, args ...interface{})
	Error(args ...interface{})
	Errorz(msg string, fields ...Field)

	Fatalf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalz(msg string, fields ...Field)

	Infof(format string, args ...interface{})
	Info(args ...interface{})
	Infoz(msg string, fields ...Field)

	Warnf(format string, args ...interface{})
	Warn(args ...interface{})
	Warnz(msg string, fields ...Field)

	Debugf(format string, args ...interface{})
	Debug(args ...interface{})
	Debugz(msg string, fields ...Field)
}

// StandardLogger initializes the standard logger
type standardLogger struct {
	logger *zap.SugaredLogger
	log    *zap.Logger
}

type lumberjackSink struct {
	*lumberjack.Logger
}

func (lumberjackSink) Sync() error {
	return nil
}

type bufwriter chan []byte

func (bw bufwriter) Write(p []byte) (int, error) {
	bw <- p
	return len(p), nil
}

var loggerPointer *standardLogger

func SetLogger(l *standardLogger) {
	loggerPointer = l
}

func GetLogger() *standardLogger {
	return loggerPointer
}

func NewBufwriter(n int, logFile string) bufwriter {
	w := make(bufwriter, n)
	logwriter := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    500, // megabytes
		MaxBackups: 10,
		MaxAge:     7, //days
	}
	go func(l *lumberjack.Logger, c bufwriter) {
		for p := range c {
			os.Stdout.Write(p)
			l.Write(p)
		}
	}(logwriter, w)
	return w
}

func getConfigFromInterface(confi interface{}) *config.Logger {
	conf := confi.(config.Logger)
	return &conf
}

// NewService initializes the standard logger
func NewService(config interface{}) *standardLogger {
	conf := getConfigFromInterface(config)

	atom := zap.NewAtomicLevel()
	atom.SetLevel(GetLevel(conf.LoggingLevel)) // level has been set

	encoderConfig := zapcore.EncoderConfig{
		MessageKey: "message",

		LevelKey:    "level",
		EncodeLevel: zapcore.CapitalLevelEncoder,

		TimeKey:    "time",
		EncodeTime: zapcore.ISO8601TimeEncoder,

		// Commented as we manually add caller
		CallerKey: "caller",

		EncodeCaller: zapcore.FullCallerEncoder,
		LineEnding:   zapcore.DefaultLineEnding,
	}

	if strings.TrimSpace(conf.LogFileName) != "" {
		ll := lumberjack.Logger{
			Filename:   conf.LogFileName,
			MaxSize:    conf.LogFileSizeCappingInMBs,
			MaxBackups: conf.MaxLogBackupsCount,
			MaxAge:     conf.MaxOldLogRetentionInDays,
			Compress:   conf.OldLogsCompressionRequired,
		}

		zap.RegisterSink("lumberjack", func(*url.URL) (zap.Sink, error) {
			return lumberjackSink{
				Logger: &ll,
			}, nil
		})
	}

	cfg := zap.Config{
		Encoding:         "json",
		Level:            atom,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig:    encoderConfig,
	}

	if strings.TrimSpace(conf.LogFileName) != "" {
		cfg.OutputPaths = append(cfg.OutputPaths, conf.LogFileName)
	}

	logger, nerr := cfg.Build() // error shouldn't happen but still I am handling it
	var sugar *zap.SugaredLogger
	if nerr != nil {
		sugar = zap.NewExample().Sugar()
		sugar.Error("Was unable to create logger file!")
		sugar.Error("Was unable to create desired logger, running on simple logger!")
	} else {
		sugar = logger.Sugar()
		sugar.Error("Was unable to create logger file!")
	}

	sugar = sugar.WithOptions(zap.AddCallerSkip(1))
	defer sugar.Sync()
	return &standardLogger{sugar, logger}
}

func (s *standardLogger) GetZapLogger() *zap.Logger {
	return s.log
}

func (s *standardLogger) GetSDLogger() *zap.SugaredLogger {
	return s.logger
}

func (s *standardLogger) GetLogger() *standardLogger {
	return s
}

func (s *standardLogger) Errorf(format string, args ...interface{}) {
	s.logger.Errorf(format, args...)
}

func (s *standardLogger) Error(args ...interface{}) {
	reponseMessage := "unknown"
	if err, ok := args[len(args)-1].(error); ok {
		errString := err.Error()
		fmt.Println("errString", errString)
		args = append(args[:len(args)-1], " ", errString)
		reponseMessage = err.Error()
	}
	s.logger.WithOptions(zap.Fields(zap.Field{
		Key:    "response_message",
		Type:   15,
		String: reponseMessage,
	})).Error(args...)
}

func (s *standardLogger) Errorz(msg string, fields ...Field) {
	s.log.Error(msg, fields...)
}

func (s *standardLogger) Fatalf(format string, args ...interface{}) {
	s.logger.Fatalf(format, args...)
}

func (s *standardLogger) Fatal(args ...interface{}) {
	s.logger.Fatal(args...)
}

func (s *standardLogger) Fatalz(msg string, fields ...Field) {
	s.log.Fatal(msg, fields...)
}

func (s *standardLogger) Infof(format string, args ...interface{}) {
	s.logger.Infof(format, args...)
}

func (s *standardLogger) Info(args ...interface{}) {
	// for _, v := range args {
	// 	if fmt.Sprintf("%T", v) == "zapcore.Field" {
	// 		x := v.(zapcore.Field)
	// 		s.logger.WithOptions(zap.Fields(zap.Field{
	// 			Key:    x.Key,
	// 			Type:   x.Type,
	// 			String: x.String,
	// 		})).Error(args...)
	// 	}

	// }
	s.logger.Info(args...)
}

func (s *standardLogger) Infoz(msg string, fields ...Field) {
	s.log.Info(msg, fields...)
}

func (s *standardLogger) Warn(args ...interface{}) {
	s.logger.Warn(args...)
}

func (s *standardLogger) Warnf(format string, args ...interface{}) {
	s.logger.Warnf(format, args...)
}

func (s *standardLogger) Warnz(msg string, fields ...Field) {
	s.log.Warn(msg, fields...)
}

func (s *standardLogger) Debugf(format string, args ...interface{}) {
	fmt.Println(format, args)
	s.logger.Debugf(format, args...)
}

func (s *standardLogger) Debug(args ...interface{}) {
	s.logger.Debug(args...)
}

func (s *standardLogger) Debugz(msg string, fields ...Field) {
	s.log.Debug(msg, fields...)
}

func (s *standardLogger) Printf(format string, args ...interface{}) {
	s.logger.Infof(format, args...)
}

func (s *standardLogger) Println(args ...interface{}) {
	s.logger.Info(args...)
	s.logger.Info("\n")
}

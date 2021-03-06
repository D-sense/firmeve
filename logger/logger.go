package logging

import (
	"fmt"
	"github.com/firmeve/firmeve/kernel/contract"
	"io"
	"os"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type (
	logger struct {
		channels channels
		config   contract.Configuration
		current  string
	}

	Level string

	internalLogger = *zap.SugaredLogger

	channels map[string]internalLogger

	writers map[string]io.Writer
)

const (
	Debug Level = `debug`
	Info        = `info`
	Warn        = `warn`
	Error       = `error`
	Fatal       = `fatal`
)

var (
	mu       sync.Mutex
	levelMap = map[Level]zapcore.Level{
		Debug: zapcore.DebugLevel,
		Info:  zapcore.InfoLevel,
		Warn:  zapcore.WarnLevel,
		Error: zapcore.ErrorLevel,
		Fatal: zapcore.FatalLevel,
	}
	channelMap = map[string]func(config contract.Configuration) io.Writer{
		`file`:    newFileChannel,
		`console`: newConsoleChannel,
	}
	consoleZapEncoder = zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
	})
	fileZapEncoder = zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
	})
)

func New(config contract.Configuration) contract.Loggable {
	return &logger{
		config:   config,
		current:  config.GetString(`default`),
		channels: make(channels, 0),
	}
}

func (l *logger) Debug(message string, context ...interface{}) {
	l.channel(l.current).Debugw(message, context...)
}

func (l *logger) Info(message string, context ...interface{}) {
	l.channel(l.current).Infow(message, context...)
}

func (l *logger) Warn(message string, context ...interface{}) {
	l.channel(l.current).Warnw(message, context...)
}

func (l *logger) Error(message string, context ...interface{}) {
	l.channel(l.current).Errorw(message, context...)
}

func (l *logger) Fatal(message string, context ...interface{}) {
	l.channel(l.current).Fatalw(message, context...)
}

// Return a new Logger instance
// But still using internal channels
func (l *logger) Channel(stack string) contract.Loggable {
	return &logger{
		config:   l.config,
		channels: l.channels,
		current:  stack,
	}
}

// Get designated channel
func (l *logger) channel(stack string) internalLogger {
	if channel, ok := l.channels[stack]; ok {
		return channel
	}

	mu.Lock()
	defer mu.Unlock()

	l.channels[stack] = factory(stack, l.config)
	return l.channels[stack]
}

// ---------------------------------------------- func --------------------------------------------------

// Default internal logger
func zapLogger(config contract.Configuration, writers writers) internalLogger {
	//zapcore.EncoderConfig{
	//	TimeKey:        "time",
	//	LevelKey:       "level",
	//	NameKey:        "logger",
	//	CallerKey:      "caller",
	//	MessageKey:     "message",
	//	StacktraceKey:  "stacktrace",
	//	LineEnding:     zapcore.DefaultLineEnding,
	//	EncodeLevel:    zapcore.LowercaseLevelEncoder,
	//	EncodeTime:     zapcore.ISO8601TimeEncoder,
	//	EncodeDuration: zapcore.StringDurationEncoder,
	//	EncodeCaller:   zapcore.FullCallerEncoder,
	//}
	cores := make([]zapcore.Core, 0)
	var zapEncoder zapcore.Encoder
	for stack, write := range writers {
		if stack == `console` {
			zapEncoder = consoleZapEncoder
		} else {
			zapEncoder = fileZapEncoder
		}

		core := zapcore.NewCore(
			zapEncoder,
			zapcore.Lock(zapcore.AddSync(write)), //writer(option)
			zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
				return lvl >= levelMap[Level(config.GetStringMap(strings.Join([]string{`channels`, stack}, `.`))[`level`].(string))]
			}),
		)

		cores = append(cores, core)
	}

	return zap.New(
		zapcore.NewTee(cores...), zap.AddCallerSkip(2), zap.AddStacktrace(zap.DebugLevel),
	).Sugar()
}

// Channel factory
func factory(stack string, config contract.Configuration) internalLogger {
	var channels writers
	switch stack {
	case `file`:
		channels = writers{stack: newFileChannel(config)}
	case `console`:
		channels = writers{stack: newConsoleChannel(config)}
	case `stack`:
		channels = newStackChannel(config)
	default:
		panic(fmt.Errorf("the logger stack %s not exists", stack))
	}

	return zapLogger(config, channels)
}

// New file channel
func newFileChannel(config contract.Configuration) io.Writer {
	return &lumberjack.Logger{
		Filename:   config.GetStringMap(`channels.file`)[`path`].(string) + "/log.log",
		MaxSize:    config.GetStringMap(`channels.file`)[`size`].(int),
		MaxBackups: config.GetStringMap(`channels.file`)[`backup`].(int),
		MaxAge:     config.GetStringMap(`channels.file`)[`age`].(int),
	}
}

// New console channel
func newConsoleChannel(config contract.Configuration) io.Writer {
	return os.Stdout
}

// New stack channel
func newStackChannel(config contract.Configuration) writers {
	stacks := config.GetStringSlice(`channels.stack`)
	existsStackMap := make(writers, 0)
	for _, stack := range stacks {
		existsStackMap[stack] = channelMap[stack](config)
	}

	return existsStackMap
}

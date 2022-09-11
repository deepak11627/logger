// go:build !windows
package log

import (
	"fmt"
	"io"
	"strings"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	timeKey           = "time"
	levelKey          = "level"
	messageKey        = "msg"
	defaultLevel      = "info"
	defaultEncoding   = "json"
	logFilePathFormat = "/logs/%s.log"
	maxsize           = 20
	maxcount          = 3
	defaultInstanceID = "instance_1"
)

// config contains configuration options for creating the logger
type config struct {
	// Can be "json" or "console" for the log output
	encoding string
	// Debug determines if debug level logs are logged. Logging levels may be modified at runtime via a Manager
	logLevel string
	// logPath specifies the path to the log file. stdout and stderr are also acceptable values
	logPath string
	// Log rotation size
	rotationSize int
	// rotationCount int
	rotationCount int
	// SyslogPort can be configured to forward logs to a syslog server
	syslogConn string
	// can be either udp or tcp
	syslogProtocol string
	// output to some writer
	out io.Writer
	// instance ID of the service
	instanceID string
}

// An Option configures a logger.
type Option func(*logger)

// WithSyslog configures the syslog server for log forwarding
func WithSyslog(conn, protocol string) Option {
	return func(log *logger) {
		log.config.syslogConn = conn
		log.config.syslogProtocol = protocol
	}
}

// WithOutput configures the logger to write to stdout also
func WithOutput(w io.Writer) Option {
	return func(log *logger) {
		log.config.out = w
	}
}

// WithLogLevel overrides the default log level
func WithLogLevel(level string) Option {
	return func(log *logger) {
		log.config.logLevel = level
	}
}

// WithEncoding overrides default encoding
func WithEncoding(ec string) Option {
	return func(log *logger) {
		log.config.encoding = defaultEncoding
	}
}

// WithLogFile overrides default log file path
func WithLogFile(filePath string) Option {
	return func(log *logger) {
		log.config.logPath = filePath
	}
}

// WithrotationSize overrides default log rotation file size
func WithRotationSize(size int) Option {
	return func(log *logger) {
		log.config.rotationSize = size
	}
}

// WithrotationCount overrides default log rotation file count
func WithRotationCount(count int) Option {
	return func(log *logger) {
		log.config.rotationCount = count
	}
}

// WithrotationCount overrides default log rotation file count
func WithInstanceID(instanceID string) Option {
	return func(log *logger) {
		log.config.instanceID = instanceID
	}
}

// logger used for loggin purpose through out the application
type logger struct {
	// Zap Sugared logger
	*zap.SugaredLogger

	// Lumberjack for rog rotation
	rotator *lumberjack.Logger

	// Allows to dynamically change the log level
	level zap.AtomicLevel

	// logging configuration
	config *config
}

// NewLogger is a wrapper create around Zap logger
func NewLogger(app, version string, opts ...Option) (*logger, error) {

	// create  logger
	l := &logger{}
	l.config = &config{
		logLevel:      defaultLevel,
		logPath:       fmt.Sprintf(logFilePathFormat, strings.ToLower(app)),
		rotationSize:  maxsize,
		rotationCount: maxcount,
		encoding:      defaultEncoding,
		instanceID:    defaultInstanceID,
	}
	// apply options
	for _, opt := range opts {
		opt(l)
	}

	zapLogLevel := getZapLogLevel(l.config.logLevel)

	// level to modify the logging level dynamically
	level := zap.NewAtomicLevelAt(zapLogLevel)

	cfg := zap.NewProductionConfig()
	cfg.DisableCaller = true

	cfg.EncoderConfig.LevelKey = levelKey
	cfg.EncoderConfig.MessageKey = messageKey
	cfg.EncoderConfig.TimeKey = timeKey
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	cfg.Encoding = l.config.encoding

	targets := []zapcore.WriteSyncer{}
	if l.config.syslogConn != "" {
		// Initialize a syslog writer
		Syslog, err := GetSyslog(l.config.syslogProtocol, l.config.syslogConn, fmt.Sprintf("%s-%s", app, version))
		if err != nil {
			return nil, err
		}
		if Syslog != nil {
			targets = append(targets, zapcore.AddSync(Syslog))
		}
	}

	if l.config.out != nil {
		targets = append(targets, zapcore.AddSync(l.config.out))
	}

	// create log rotator
	var lr *lumberjack.Logger
	if l.config.encoding == defaultEncoding {
		lr = getLogRotator(l.config)
		targets = append(targets, zapcore.AddSync(lr))
	}

	// create multi write syncer for configured targets
	syncer := zapcore.NewMultiWriteSyncer(targets...)

	sl, err := cfg.Build(
		withWrapCore(syncer, cfg, level),
		zap.Fields(
			zap.String("service", app),
			zap.String("version", version),
			zap.String("instance_id", l.config.instanceID),
		),
	)
	if err != nil {
		return nil, err
	}
	defer sl.Sync()
	l.SugaredLogger = sl.Sugar()
	l.level = level
	if lr != nil {
		l.rotator = lr
	}
	return l, nil
}

// withWrapCore replaces existing Core with new, that writes to passed WriteSyncer.
func withWrapCore(ws zapcore.WriteSyncer, conf zap.Config, level zap.AtomicLevel) zap.Option {
	var enc zapcore.Encoder
	switch conf.Encoding {
	case "json":
		enc = zapcore.NewJSONEncoder(conf.EncoderConfig)
	case "console":
		enc = zapcore.NewConsoleEncoder(conf.EncoderConfig)
	default:
		panic("unknown encoding")
	}

	return zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.NewCore(enc, ws, level)
	})
}

// returns a lumberjack logger instance
func getLogRotator(conf *config) *lumberjack.Logger {
	return &lumberjack.Logger{
		Filename:   conf.logPath,
		MaxSize:    conf.rotationSize,  // MB
		MaxBackups: conf.rotationCount, // number of backups
		LocalTime:  true,
		Compress:   false, // disabled by default
	}
}

func getZapLogLevel(level string) zapcore.Level {
	var logLevel zapcore.Level
	switch level {
	case "debug":
		logLevel = zap.DebugLevel
	case "info":
		logLevel = zap.InfoLevel
	case "warn":
		logLevel = zap.WarnLevel
	case "error":
		logLevel = zap.ErrorLevel
	case "panic":
		logLevel = zap.PanicLevel
	case "fatal":
		logLevel = zap.FatalLevel
	default:
		panic(fmt.Sprintf("unsupported log level %s. debug, info, warn, error, panic and fatal are the only supported loglevels.", level))
	}
	return logLevel
}

// Debug logs a debug message to the zap logger
func (l *logger) Debug(msg string, keyvals ...interface{}) {
	l.Debugw(msg, keyvals...)
}

// Info logs an info message to the zap logger
func (l *logger) Info(msg string, keyvals ...interface{}) {
	l.Infow(msg, keyvals...)
}

// Warn logs a warning message to the zap logger
func (l *logger) Warn(msg string, keyvals ...interface{}) {
	l.Warnw(msg, keyvals...)
}

// Error logs a error message to the zap logger
func (l *logger) Error(msg string, keyvals ...interface{}) {
	l.Errorw(msg, keyvals...)
}

// Panic logs a panic message to the zap logger
func (l *logger) Panic(msg string, keyvals ...interface{}) {
	l.Panicw(msg, keyvals...)
}

// Fatal logs a fatal message to the zap logger
func (l *logger) Fatal(msg string, keyvals ...interface{}) {
	l.Fatalw(msg, keyvals...)
}

// Debugf logs a formatted debug message
func (l *logger) Debugf(format string, args ...interface{}) {
	l.SugaredLogger.Debugf(format, args...)
}

// Printf logs a formatted  message
func (l *logger) Printf(format string, args ...interface{}) {
	l.SugaredLogger.Infof(format, args...)
}

// With returns a new logger with the provided keyvals added to its context
func (l *logger) With(keyvals ...interface{}) Logger {
	return &logger{l.SugaredLogger.With(keyvals...), nil, l.level, l.config}
}

// SetLogLevel update the current logging level to the supplied one
func (l *logger) SetLogLevel(level string) {
	switch level {
	case "debug":
		l.level.SetLevel(zap.DebugLevel)
	case "info":
		l.level.SetLevel(zap.InfoLevel)
	case "warn":
		l.level.SetLevel(zap.WarnLevel)
	case "error":
		l.level.SetLevel(zap.ErrorLevel)
	case "panic":
		l.level.SetLevel(zap.PanicLevel)
	case "fatal":
		l.level.SetLevel(zap.FatalLevel)
	default:
		l.Warn("can't change to a unsupported log level. Only debug, info, warn, error, panic and fatal levels are supported.")
	}
}

// SetLogRotationSize update the current log rotation file size to the supplied one
func (l *logger) SetLogRotationSize(size int) {
	l.rotator.MaxSize = size
}

// SetLogRotationCount update the current log rotation file count to the supplied one
func (l *logger) SetLogRotationCount(count int) {
	l.rotator.MaxBackups = count
}

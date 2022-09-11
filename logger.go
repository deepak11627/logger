package log

// A common logger interface to be used throughout  platfrom components
type Logger interface {
	Debug(msg string, keyvals ...interface{})

	Info(msg string, keyvals ...interface{})

	Warn(msg string, keyvals ...interface{})

	Error(msg string, keyvals ...interface{})

	Panic(msg string, keyvals ...interface{})

	Fatal(msg string, keyvals ...interface{})

	Printf(format string, args ...interface{})

	Debugf(format string, args ...interface{})

	With(keyvals ...interface{}) Logger

	// Set log level
	SetLogLevel(level string)
	// Set log rotation size
	SetLogRotationSize(size int)
	// Set log rotation count
	SetLogRotationCount(count int)
}

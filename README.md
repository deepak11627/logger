## Logger
 This logger is common logger package designed to be used in any Go project. With this wrapper one can do following easily,
 - Log rotation
 - Sending logs to multiple syncer such as rsyslog
 - Custom service properties such as service name, version etc.

## Usage
Follow the instructions to use  logger.

### Instantiation
To create a logger instance one only needs the application name and version, two params to be supplied to the log.NewLogger() method. In this case all other values will be set to default. 
To override default configuration just use methods such as logger.WithLogLevel("debug").
```
	l, err := log.NewLogger("MyService", "1.0.0", logger.WithLogLevel("debug"), logger.WithRotationSize(2))
	if err != nil {
		log.Fatal("failed to get the logger, error: ", err)
	}
```

### Log levels and Logging
Supported log levels are `debug`, `info`, `warn`, `error`, `panic`, and `fatal`.

Logging with different methods,
```
l, err := log.NewLogger("MyService", "1.0.0")
if err != nil {
    log.Fatal("failed to get the logger, error: ", err)
}
l.Debug("my logs message goes here")
l.Info("log message with addtional fields", "field", "value")
```
Context based logging,
```
l, err := log.NewLogger("MyService", "1.0.0")
if err != nil {
    log.Fatal("failed to get the logger, error: ", err)
}
ctxLogger := l.With("myfield", "myvalue")
ctxLogger.Debug("my message goes here")
ctxLogger.Debug("my another message goes here")
// both above logs will have myfield and myvalue been set to the log entry.
```

### Change Log level
```
l, err := log.NewLogger("MyService", "1.0.0")
if err != nil {
    log.Fatal("failed to get the logger, error: ", err)
}
l.SetLogLevel("warn")
// use any one of the support log level in the argument.
```
### Change log rotation params
```
l, err := log.NewLogger("MyService", "1.0.0")
if err != nil {
    log.Fatal("failed to get the logger, error: ", err)
}
l.SetRotationSize(20)
// Specify an integer value for size in MB.
l.SetRotationCount(5)
// Specify an integer value for total number files to be kept on disk
```

### Send logs to Syslog
```
l, err := log.NewLogger(&logger.Config{SyslogPort: "514",})
if err != nil {
    log.Fatal("failed to get the logger, error: ", err)
}
```
package pram

type (
	// Logger represents a logger
	Logger interface {
		Printf(format string, a ...interface{})
	}

	noopLogger struct{}
)

var logger Logger = new(noopLogger)

// SetLogger sets the logger
func SetLogger(l Logger) {
	if l == nil {
		l = new(noopLogger)
	}
	logger = l
}

// Logf writes the input to the configured logger
func Logf(format string, a ...interface{}) {
	logger.Printf(format, a...)
}

// Printf is a noop
func (l *noopLogger) Printf(format string, a ...interface{}) {}

package pram

type (
	// Logger represents a logger
	Logger interface {
		Print(v ...interface{})
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

// Log logs the input to the configured logger
func Log(v ...interface{}) {
	logger.Print(v...)
}

// Logf logs the input to the configured logger
func Logf(format string, a ...interface{}) {
	logger.Printf(format, a...)
}

func (l *noopLogger) Print(v ...interface{}) {}

func (l *noopLogger) Printf(format string, a ...interface{}) {}

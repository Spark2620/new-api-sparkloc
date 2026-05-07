package common

const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)

func currentLogSeverity() int {
	switch LogLevel {
	case LogLevelDebug:
		return 10
	case LogLevelInfo:
		return 20
	case LogLevelWarn:
		return 30
	case LogLevelError:
		return 40
	default:
		return 20
	}
}

func LogLevelEnabled(level string) bool {
	var target int
	switch level {
	case LogLevelDebug:
		target = 10
	case LogLevelInfo:
		target = 20
	case LogLevelWarn:
		target = 30
	case LogLevelError:
		target = 40
	default:
		target = 20
	}
	return target >= currentLogSeverity()
}

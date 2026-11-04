package log

type Logger interface {
	Error(string, ...any)
	Info(string, ...any)
	Debug(string, ...any)
	Warn(string, ...any)
}

package controller

type Logger interface {
	Info(args ...interface{})
	Error(args ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
}

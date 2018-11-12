package transactional

type Configuration struct {
	Name   string
	Logger Logger
}

type Logger interface {
	Printf(format string, v ...interface{})
}

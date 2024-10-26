package logger

type LoggerFrom string

const (
	General    LoggerFrom = "general"
	AppMetrics LoggerFrom = "app-metrics"
	GinServer  LoggerFrom = "gin-server"
	CStorage   LoggerFrom = "cassandra-storage"
)

type Logger interface {
	Debug(lf LoggerFrom, msg string, extra map[string]interface{})
	Info(lf LoggerFrom, msg string, extra map[string]interface{})
	Warn(lf LoggerFrom, msg string, extra map[string]interface{})
	Error(lf LoggerFrom, msg string, extra map[string]interface{})
	Fatal(lf LoggerFrom, msg string, extra map[string]interface{})
}

package ginserver

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yasseralhendawy/hexagonal_chat/pkg/logger/logger"
	appmetrics "github.com/yasseralhendawy/hexagonal_chat/pkg/metrics/adapter"
)

type GinServer struct {
	Engin *gin.Engine
}

type GinOpt func(*GinServer)

func InitServer(opts ...GinOpt) *GinServer {
	gin.SetMode(gin.ReleaseMode)
	api := &GinServer{
		Engin: gin.New(),
	}
	for _, opt := range opts {
		opt(api)
	}
	return api
}

func Logger(lg logger.Logger) GinOpt {
	return func(s *GinServer) {
		s.Engin.Use(handleLogger(lg))
	}
}

func Metric(metric appmetrics.Metrics) GinOpt {
	return func(s *GinServer) {
		s.Engin.Use(handleMetrics(metric))
	}
}

// func (api *ginFramework) addGroup(add string) *gin.RouterGroup {
// 	return api.engin.Group(add)
// }

func (api *GinServer) Run(addr ...string) error {
	return api.Engin.Run(addr...)
}

func handleLogger(l logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()
		end := time.Now()

		latency := end.Sub(start)
		keys := map[string]interface{}{}
		keys["path"] = path
		keys["method"] = c.Request.Method
		keys["query"] = query
		keys["ip"] = c.ClientIP()
		keys["Latency"] = latency
		keys["status"] = c.Writer.Status()
		keys["user-agent"] = c.Request.UserAgent()
		keys["error-message"] = c.Errors.ByType(gin.ErrorTypePrivate).String()

		l.Info(logger.GinServer, "", keys)

	}
}

func handleMetrics(m appmetrics.Metrics) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		path := ctx.FullPath()
		method := ctx.Request.Method
		ctx.Next()
		status := ctx.Writer.Status()
		m.LatencyWithLabelValues(float64(time.Since(start)/time.Millisecond), path, method, strconv.Itoa(status))
	}
}

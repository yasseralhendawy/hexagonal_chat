package ginserver

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yasseralhendawy/hexagonal_chat/pkg/claims"
	logger "github.com/yasseralhendawy/hexagonal_chat/pkg/logger/adapter"
	appmetrics "github.com/yasseralhendawy/hexagonal_chat/pkg/metrics/adapter"
	gorillasocket "github.com/yasseralhendawy/hexagonal_chat/pkg/websocket/gorilla_socket"
)

const claimKey = "claim"

type GinServer struct {
	Engin            *gin.Engine
	websocketManager *gorillasocket.Server
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

func WebSocket(ws *gorillasocket.Server) GinOpt {
	return func(s *GinServer) {
		s.websocketManager = ws
	}
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
func Auth(tokenization *claims.TokenGenerator) GinOpt {
	return func(gs *GinServer) {
		gs.Engin.Use(handleClaims(tokenization))
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

func handleClaims(tokenization *claims.TokenGenerator) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		header := ctx.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid Authorization header"})
			return
		}
		token := strings.TrimPrefix(header, "Bearer ")
		claims, err := tokenization.ValidateToken(token)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}
		ctx.Set(claimKey, claims)
	}
}

func getClaims(ctx *gin.Context) (*claims.ClaimsData, error) {
	c, ok := ctx.Get(claimKey)
	if !ok {
		return nil, errors.New("missing claims")
	}
	claimsData, ok := c.(*claims.ClaimsData)
	if !ok {
		return nil, errors.New("invalid claims type")
	}
	return claimsData, nil
}

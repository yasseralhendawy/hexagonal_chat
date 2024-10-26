package main

import (
	"github.com/yasseralhendawy/hexagonal_chat/config"
	"github.com/yasseralhendawy/hexagonal_chat/domain/auth"
	authapp "github.com/yasseralhendawy/hexagonal_chat/internal/application/auth"
	cstorage "github.com/yasseralhendawy/hexagonal_chat/internal/infrastructure/cassandra"
	ginhs "github.com/yasseralhendawy/hexagonal_chat/internal/presentation/gin_handlers/auth"
	"github.com/yasseralhendawy/hexagonal_chat/pkg/claims"
	"github.com/yasseralhendawy/hexagonal_chat/pkg/logger/logger"
	zaplogger "github.com/yasseralhendawy/hexagonal_chat/pkg/logger/zap_logger"
	prometrics "github.com/yasseralhendawy/hexagonal_chat/pkg/metrics/prometheus"
	ginserver "github.com/yasseralhendawy/hexagonal_chat/pkg/servers/gin_server"
)

func main() {
	//lets get the configurations
	cfg, err := config.GetConfig("/exp1")
	if err != nil {
		return
	}
	//lets create common pkgs that might be used in different layers.
	//1- logger
	lg := zaplogger.CreateLogger(&cfg.Log)

	//2- metrics
	metric, err := prometrics.CreateMetrics()
	if err != nil {
		lg.Fatal(logger.AppMetrics, err.Error(), nil)
		return
	}
	//lets create our database and establish the session
	db, err := cstorage.NewCassandraSession(cfg.Cassandra)
	if err != nil {
		lg.Fatal(logger.CStorage, err.Error(), nil)
		return
	}
	// this function will also ensure that the table is established
	repo, err := db.NewAuthRepo(metric)
	if err != nil {
		lg.Error(logger.CStorage, err.Error(), nil)
		return
	}
	defer db.StopSession()
	// let's rap all things togeher
	handler := ginhs.AuthHandler{
		Server: ginserver.InitServer(ginserver.Logger(lg), ginserver.Metric(metric)),
		App: &authapp.App{
			Tokenization:  claims.NewTokenGenerator(&cfg.Jwt),
			DomainService: auth.New(repo),
		},
	}
	lg.Info(logger.GinServer, "Server Started Successfuly", nil)
	handler.Run(":8081")
}

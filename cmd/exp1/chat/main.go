package main

import (
	"github.com/yasseralhendawy/hexagonal_chat/config"
	"github.com/yasseralhendawy/hexagonal_chat/domain/chat"
	"github.com/yasseralhendawy/hexagonal_chat/domain/user"
	chatmessage "github.com/yasseralhendawy/hexagonal_chat/internal/application/chat_message"
	userchat "github.com/yasseralhendawy/hexagonal_chat/internal/application/user_chat"
	cstorage "github.com/yasseralhendawy/hexagonal_chat/internal/infrastructure/cassandra"
	ginserver "github.com/yasseralhendawy/hexagonal_chat/internal/presentation/gin_server"
	logger "github.com/yasseralhendawy/hexagonal_chat/pkg/logger/adapter"
	zaplogger "github.com/yasseralhendawy/hexagonal_chat/pkg/logger/zap_logger"
	prometrics "github.com/yasseralhendawy/hexagonal_chat/pkg/metrics/prometheus"
)

func main() {
	//lets get the configrations for now just check the error
	cfg, err := config.GetConfig("/exp1")
	if err != nil {
		panic(err)
	}
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
	cm_repo, err := db.NewChatRepo(metric)
	if err != nil {
		lg.Error(logger.CStorage, err.Error(), nil)
		return
	}
	uc_repo, err := db.NewUserChatRepo(metric)
	if err != nil {
		lg.Error(logger.CStorage, err.Error(), nil)
		return
	}
	defer db.StopSession()

	handler := ginserver.ChatHandler{
		Server: ginserver.InitServer(ginserver.Logger(lg), ginserver.Metric(metric)),
		CMApp: &chatmessage.App{
			Domain: &chat.ChatService{Storage: cm_repo},
		},
		UCApp: &userchat.App{
			Domain: &user.Service{Storage: uc_repo},
		},
	}
	lg.Info(logger.GinServer, "Server Started Successfuly", nil)
	handler.Run(":8082")
}

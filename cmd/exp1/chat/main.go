package main

import (
	"github.com/yasseralhendawy/hexagonal_chat/config"
	logger "github.com/yasseralhendawy/hexagonal_chat/pkg/logger/adapter"
	zaplogger "github.com/yasseralhendawy/hexagonal_chat/pkg/logger/zap_logger"
)

func main() {
	//lets get the configrations for now just check the error
	cfg, err := config.GetConfig("/exp1")
	if err != nil {
		panic(err)
	}
	lg := zaplogger.CreateLogger(&cfg.Log)
	lg.Info(logger.General, "Hello world from chat", nil)
}

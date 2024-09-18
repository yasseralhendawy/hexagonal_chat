package main

import (
	"github.com/yasseralhendawy/hexagonal_chat/config"
	"github.com/yasseralhendawy/hexagonal_chat/pkg/logger/logger"
	zaplogger "github.com/yasseralhendawy/hexagonal_chat/pkg/logger/zap_logger"
	prometrics "github.com/yasseralhendawy/hexagonal_chat/pkg/metrics/prometheus"
)

func main() {
	//lets get the configrations for now just check the error
	cfg, err := config.GetConfig("/exp1")
	if err != nil {
		panic(err)
	}
	lg := zaplogger.CreateLogger(&cfg.Log)
	lg.Info(logger.General, "Hello world from auth", nil)

	_, err = prometrics.CreateMetrics()
	if err != nil {
		lg.Error(logger.AppMetrics, err.Error(), nil)
	}
}

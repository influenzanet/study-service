package main

import (
	"context"
	"log"

	"github.com/coneno/logger"
	"github.com/influenzanet/study-service/internal/config"
	"github.com/influenzanet/study-service/pkg/dbs/globaldb"
	"github.com/influenzanet/study-service/pkg/dbs/studydb"
	gc "github.com/influenzanet/study-service/pkg/grpc/clients"
	"github.com/influenzanet/study-service/pkg/grpc/service"
	"github.com/influenzanet/study-service/pkg/studytimer"
	"github.com/influenzanet/study-service/pkg/types"
)

func main() {
	conf := config.InitConfig()

	logger.SetLevel(conf.LogLevel)

	studyDBService := studydb.NewStudyDBService(conf.StudyDBConfig)
	globalDBService := globaldb.NewGlobalDBService(conf.GlobalDBConfig)

	sTimerService := studytimer.NewStudyTimerService(conf.Study, studyDBService, globalDBService)
	sTimerService.Run()

	clients := &types.APIClients{}

	loggingClient, close := gc.ConnectToLoggingService(conf.ServiceURLs.LoggingService)
	defer close()
	clients.LoggingService = loggingClient

	ctx := context.Background()
	if err := service.RunServer(
		ctx,
		conf.Port,
		clients,
		studyDBService,
		globalDBService,
		conf.Study.GlobalSecret,
	); err != nil {
		log.Fatal(err)
	}
}

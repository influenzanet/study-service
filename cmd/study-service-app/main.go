package main

import (
	"context"

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

	ensureDBIndexes(globalDBService, studyDBService)

	sTimerService := studytimer.NewStudyTimerService(conf.Study, studyDBService, globalDBService, conf.ExternalServices, conf.Study.GlobalSecret)
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
		conf.MaxMsgSize,
		conf.PersistentStoreConfig,
		conf.ExternalServices,
	); err != nil {
		logger.Error.Fatal(err)
	}
}

func ensureDBIndexes(gdb *globaldb.GlobalDBService, sdb *studydb.StudyDBService) {
	instances, err := gdb.GetAllInstances()
	if err != nil {
		logger.Error.Printf("unexpected error when fetching instances: %v", err)
	}

	for _, i := range instances {
		sdb.CreateSurveyDefintionIndexForAllStudies(i.InstanceID)
		sdb.CreateMessageScheduledForIndexForAllStudies(i.InstanceID)
		// TODO: ensure other indexes as well
	}

}

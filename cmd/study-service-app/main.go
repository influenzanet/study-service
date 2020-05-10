package main

import (
	"context"
	"log"

	"github.com/influenzanet/study-service/internal/config"
	"github.com/influenzanet/study-service/pkg/dbs/globaldb"
	"github.com/influenzanet/study-service/pkg/dbs/studydb"
	"github.com/influenzanet/study-service/pkg/service"
	"github.com/influenzanet/study-service/pkg/studytimer"
)

func main() {
	conf := config.InitConfig()
	studyDBService := studydb.NewStudyDBService(conf.StudyDBConfig)
	globalDBService := globaldb.NewGlobalDBService(conf.GlobalDBConfig)

	sTimerService := studytimer.NewStudyTimerService(conf.Study, studyDBService, globalDBService)
	sTimerService.Run()

	ctx := context.Background()
	if err := service.RunServer(
		ctx,
		conf.Port,
		studyDBService,
		globalDBService,
		conf.Study.GlobalSecret,
	); err != nil {
		log.Fatal(err)
	}
}

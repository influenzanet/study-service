package service

import (
	"github.com/influenzanet/study-service/pkg/dbs/globaldb"
	"github.com/influenzanet/study-service/pkg/dbs/studydb"
)

const (
	// apiVersion is version of API is provided by server
	apiVersion = "v1"
)

type studyServiceServer struct {
	// clients           *types.APIClients
	studyDBservice    *studydb.StudyDBService
	globalDBService   *globaldb.GlobalDBService
	StudyGlobalSecret string
}

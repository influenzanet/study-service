package studydb

import (
	"context"
	"time"

	"github.com/coneno/logger"
	"github.com/influenzanet/study-service/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type StudyDBService struct {
	DBClient     *mongo.Client
	timeout      int
	DBNamePrefix string
}

func NewStudyDBService(configs types.DBConfig) *StudyDBService {
	var err error
	dbClient, err := mongo.NewClient(
		options.Client().ApplyURI(configs.URI),
		options.Client().SetMaxConnIdleTime(time.Duration(configs.IdleConnTimeout)*time.Second),
		options.Client().SetMaxPoolSize(configs.MaxPoolSize),
	)
	if err != nil {
		logger.Error.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(configs.Timeout)*time.Second)
	defer cancel()

	err = dbClient.Connect(ctx)
	if err != nil {
		logger.Error.Fatal(err)
	}

	ctx, conCancel := context.WithTimeout(context.Background(), time.Duration(configs.Timeout)*time.Second)
	err = dbClient.Ping(ctx, nil)
	defer conCancel()
	if err != nil {
		logger.Error.Fatal("fail to connect to DB: " + err.Error())
	}

	return &StudyDBService{
		DBClient:     dbClient,
		timeout:      configs.Timeout,
		DBNamePrefix: configs.DBNamePrefix,
	}
}

// Collections
func (dbService *StudyDBService) collectionRefStudyInfos(instanceID string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.DBNamePrefix + instanceID + "_studyDB").Collection("study-infos")
}

func (dbService *StudyDBService) collectionRefStudyParticipant(instanceID string, studyKey string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.DBNamePrefix + instanceID + "_studyDB").Collection(studyKey + "_participants")
}

func (dbService *StudyDBService) collectionRefStudySurveys(instanceID string, studyKey string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.DBNamePrefix + instanceID + "_studyDB").Collection(studyKey + "_surveys")
}

func (dbService *StudyDBService) collectionRefReportHistory(instanceID string, studyKey string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.DBNamePrefix + instanceID + "_studyDB").Collection(studyKey + "_reports")
}

func (dbService *StudyDBService) collectionRefSurveyResponses(instanceID string, studyKey string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.DBNamePrefix + instanceID + "_studyDB").Collection(studyKey + "_surveyResponses")
}

func (dbService *StudyDBService) collectionRefParticipantFiles(instanceID string, studyKey string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.DBNamePrefix + instanceID + "_studyDB").Collection(studyKey + "_participantFiles")
}

func (dbService *StudyDBService) collectionRefConfidentialResponses(instanceID string, studyKey string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.DBNamePrefix + instanceID + "_studyDB").Collection(studyKey + "_confidentialResponses")
}

func (dbService *StudyDBService) collectionRefResearcherMessages(instanceID string, studyKey string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.DBNamePrefix + instanceID + "_studyDB").Collection(studyKey + "_researcherMessages")
}

func (dbService *StudyDBService) collectionRefStudyRules(instanceID string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.DBNamePrefix + instanceID + "_studyDB").Collection("studyRules")
}

// DB utils
func (dbService *StudyDBService) getContext() (ctx context.Context, cancel context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Duration(dbService.timeout)*time.Second)
}

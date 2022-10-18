package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/influenzanet/study-service/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SurveyModelDBMigrator struct {
	DBClient     *mongo.Client
	timeout      int
	DBNamePrefix string
}

func getStudyDBConfig() types.DBConfig {
	connStr := os.Getenv("STUDY_DB_CONNECTION_STR")
	username := os.Getenv("STUDY_DB_USERNAME")
	password := os.Getenv("STUDY_DB_PASSWORD")
	prefix := os.Getenv("STUDY_DB_CONNECTION_PREFIX") // Used in test mode
	if connStr == "" || username == "" || password == "" {
		log.Fatal("Couldn't read DB credentials.")
	}
	URI := fmt.Sprintf(`mongodb%s://%s:%s@%s`, prefix, username, password, connStr)

	var err error
	Timeout, err := strconv.Atoi(os.Getenv("DB_TIMEOUT"))
	if err != nil {
		log.Fatal("DB_TIMEOUT: " + err.Error())
	}
	IdleConnTimeout, err := strconv.Atoi(os.Getenv("DB_IDLE_CONN_TIMEOUT"))
	if err != nil {
		log.Fatal("DB_IDLE_CONN_TIMEOUT" + err.Error())
	}
	mps, err := strconv.Atoi(os.Getenv("DB_MAX_POOL_SIZE"))
	MaxPoolSize := uint64(mps)
	if err != nil {
		log.Fatal("DB_MAX_POOL_SIZE: " + err.Error())
	}

	DBNamePrefix := os.Getenv("DB_DB_NAME_PREFIX")

	return types.DBConfig{
		URI:             URI,
		Timeout:         Timeout,
		IdleConnTimeout: IdleConnTimeout,
		MaxPoolSize:     MaxPoolSize,
		DBNamePrefix:    DBNamePrefix,
	}
}

func NewSurveyModelDBMigrator(configs types.DBConfig) *SurveyModelDBMigrator {
	var err error
	dbClient, err := mongo.NewClient(
		options.Client().ApplyURI(configs.URI),
		options.Client().SetMaxConnIdleTime(time.Duration(configs.IdleConnTimeout)*time.Second),
		options.Client().SetMaxPoolSize(configs.MaxPoolSize),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(configs.Timeout)*time.Second)
	defer cancel()

	err = dbClient.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	ctx, conCancel := context.WithTimeout(context.Background(), time.Duration(configs.Timeout)*time.Second)
	err = dbClient.Ping(ctx, nil)
	defer conCancel()
	if err != nil {
		log.Fatal("fail to connect to DB: " + err.Error())
	}

	return &SurveyModelDBMigrator{
		DBClient:     dbClient,
		timeout:      configs.Timeout,
		DBNamePrefix: configs.DBNamePrefix,
	}
}

func (dbService *SurveyModelDBMigrator) collectionRefStudySurveys(instanceID string, studyKey string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.DBNamePrefix + instanceID + "_studyDB").Collection(studyKey + "_surveys")
}

func (dbService *SurveyModelDBMigrator) collectionRefStudySurveysOldBackup(instanceID string, studyKey string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.DBNamePrefix + instanceID + "_studyDB").Collection(studyKey + "_surveys_oldModelBackup")
}

// DB utils
func (dbService *SurveyModelDBMigrator) getContext() (ctx context.Context, cancel context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Duration(dbService.timeout)*time.Second)
}

func (dbService *SurveyModelDBMigrator) FindAllOldSurveyDefs(instanceID string, studyKey string) (surveys []OldSurvey, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{
		"current": bson.M{
			"$exists": true,
		},
	}

	var opts *options.FindOptions
	cur, err := dbService.collectionRefStudySurveys(instanceID, studyKey).Find(
		ctx,
		filter,
		opts,
	)

	if err != nil {
		return surveys, err
	}
	defer cur.Close(ctx)

	surveys = []OldSurvey{}
	for cur.Next(ctx) {
		var result OldSurvey
		err := cur.Decode(&result)
		if err != nil {
			return surveys, err
		}
		surveys = append(surveys, result)
	}
	if err := cur.Err(); err != nil {
		return surveys, err
	}

	return surveys, nil
}

func (dbService *SurveyModelDBMigrator) SaveOldSurveysIntoBackup(instanceID string, studyKey string, surveys []OldSurvey) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	docs := []interface{}{}
	for _, s := range surveys {
		docs = append(docs, s)
	}
	_, err := dbService.collectionRefStudySurveysOldBackup(instanceID, studyKey).InsertMany(ctx, docs, nil)
	return err
}

func (dbService *SurveyModelDBMigrator) DeleteOldSurveyByKey(instanceID string, studyKey string, surveyKey string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{"current.surveyDefinition.key": surveyKey}
	_, err := dbService.collectionRefStudySurveys(instanceID, studyKey).DeleteMany(ctx, filter, nil)
	return err
}

func (dbService *SurveyModelDBMigrator) SaveNewSurveyHistory(instanceID string, studyKey string, surveys []*types.Survey) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	docs := []interface{}{}
	for _, s := range surveys {
		// emptry object id, to let mongo create a new one, and avoid collisions
		s.ID = primitive.ObjectID{}
		docs = append(docs, s)
	}
	_, err := dbService.collectionRefStudySurveys(instanceID, studyKey).InsertMany(ctx, docs, nil)
	return err
}

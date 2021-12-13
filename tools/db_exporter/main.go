package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/coneno/logger"
	"github.com/influenzanet/study-service/pkg/dbs/studydb"
	"github.com/influenzanet/study-service/pkg/types"
)

var (
	studyDB   *studydb.StudyDBService
	outputDir string
)

func init() {
	conf := getStudyDBConfig()
	studyDB = studydb.NewStudyDBService(conf)

	outputDir = "exports"
	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		logger.Error.Fatal(err)
	}
}

const indentWith = "" // "  "

func main() {
	instanceID := os.Getenv("INSTANCE_ID")
	studyKey := os.Getenv("STUDY_KEY")
	surveyKey := os.Getenv("SURVEY_KEY")

	downloadSurveyDef(instanceID, studyKey, surveyKey)
	downloadResponses(instanceID, studyKey, surveyKey)
}

func downloadSurveyDef(instanceID, studyKey, surveyKey string) {
	surveyDef, err := studyDB.FindSurveyDef(instanceID, studyKey, surveyKey)
	if err != nil {
		logger.Error.Fatal(err)
	}

	file, _ := json.MarshalIndent(surveyDef, "", indentWith)

	fileName := outputDir + "/surveyDef.json"
	err = ioutil.WriteFile(fileName, file, 0644)
	if err != nil {
		logger.Error.Println(err)
	}
}

func downloadResponses(instanceID, studyKey, surveyKey string) {
	responses := []types.SurveyResponse{}

	err := studyDB.PerformActionForSurveyResponses(context.Background(), instanceID, studyKey,
		surveyKey,
		0, 0, func(instanceID, studyKey string, response types.SurveyResponse, args ...interface{}) error {
			responses = append(responses, response)
			return nil
		})
	if err != nil {
		logger.Error.Fatal(err)
	}

	file, _ := json.MarshalIndent(responses, "", indentWith)

	fileName := outputDir + "/responses.json"
	_ = ioutil.WriteFile(fileName, file, 0644)
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

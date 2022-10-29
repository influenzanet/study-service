package main

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/coneno/logger"
	"github.com/influenzanet/study-service/pkg/types"
)

func main() {
	conversionMode := flag.String("mode", "JSON", "what do you want to convert with the tool (JSON or DB or DBtest)")

	inputJSON := flag.String("input", "", "path and name of the input file that should be converted")
	instanceIDPtr := flag.String("instanceID", "", "instanceID of a specific instance to migrate surveys for")
	studyKeyPtr := flag.String("studyKey", "", "studyKey of a specific study to migrate surveys for")

	flag.Parse()
	switch *conversionMode {
	case "JSON":
		if len(*inputJSON) < 1 {
			logger.Error.Fatal("missing 'input'. Use -input=/path/to/file.json to define the file to be converted.")
		}
		oldSurveyHistory := readOldSurveyFromJSON(*inputJSON)
		saveNewSurveyIntoJSON(oldSurveyHistory.ToNew(), "newSurveyHistory.json")
	case "DB":
		instanceID := *instanceIDPtr
		studyKey := *studyKeyPtr
		if len(instanceID) < 1 {
			logger.Error.Fatal("instanceID missing")
		}
		if len(studyKey) < 1 {
			logger.Error.Fatal("studyKey missing")
		}
		dbConfig := getStudyDBConfig()
		handleDBMigration(instanceID, studyKey, dbConfig)
	case "DBtest":
		instanceID := *instanceIDPtr
		studyKey := *studyKeyPtr
		if len(instanceID) < 1 {
			logger.Error.Fatal("instanceID missing")
		}
		if len(studyKey) < 1 {
			logger.Error.Fatal("studyKey missing")
		}
		dbConfig := getStudyDBConfig()
		testDBMigration(instanceID, studyKey, dbConfig)
	default:
		logger.Error.Fatalf("unknown conversion mode: %s", *conversionMode)
	}
}

func handleDBMigration(instanceID string, studyKey string, dbConfig types.DBConfig) {
	dbMigrator := NewSurveyModelDBMigrator(dbConfig)

	oldSurveys, err := dbMigrator.FindAllOldSurveyDefs(instanceID, studyKey)
	if err != nil {
		logger.Error.Fatal(err)
	}

	if len(oldSurveys) < 1 {
		logger.Error.Fatalf("no old surveys found in %s:%s", instanceID, studyKey)
	}

	err = dbMigrator.SaveOldSurveysIntoBackup(instanceID, studyKey, oldSurveys)
	if err != nil {
		logger.Error.Fatalf("unexpected error when creating backup: %v", err)
	}

	for _, oldSurveyHistory := range oldSurveys {
		newSurveyHistory := oldSurveyHistory.ToNew()
		err = dbMigrator.SaveNewSurveyHistory(instanceID, studyKey, newSurveyHistory)
		if err != nil {
			logger.Error.Fatalf("stopping due to unexpected error during saving history for %s: %v", oldSurveyHistory.Current.SurveyDefinition.Key, err)
			continue
		}
		err = dbMigrator.DeleteOldSurveyByKey(instanceID, studyKey, oldSurveyHistory.Current.SurveyDefinition.Key)
		if err != nil {
			logger.Error.Printf("%s: %v", oldSurveyHistory.Current.SurveyDefinition.Key, err)
		}
		logger.Info.Printf("%s migrated", oldSurveyHistory.Current.SurveyDefinition.Key)
	}
}

func testDBMigration(instanceID string, studyKey string, dbConfig types.DBConfig) {
	logger.Info.Printf("Testing survey definition DB model migration for %s:%s", instanceID, studyKey)
	logger.Info.Println("This run does not modify the database. Run the tool with -mode=DB to apply the changes.")
	dbMigrator := NewSurveyModelDBMigrator(dbConfig)

	oldSurveys, err := dbMigrator.FindAllOldSurveyDefs(instanceID, studyKey)
	if err != nil {
		logger.Error.Fatal(err)
	}
	if len(oldSurveys) < 1 {
		logger.Error.Fatalf("no old surveys found in %s:%s", instanceID, studyKey)
	}

	prevCount := len(oldSurveys)
	newCount := 0
	for _, oldSurveyHistory := range oldSurveys {
		newSurveyHistory := oldSurveyHistory.ToNew()
		newCount += len(newSurveyHistory)
		logger.Info.Printf("%s will be converted into %d documents.", oldSurveyHistory.Current.SurveyDefinition.Key, len(newSurveyHistory))
	}
	logger.Info.Printf("Number of documents before update: %d", prevCount)
	logger.Info.Printf("Number of documents after update: %d", newCount)
}

func readOldSurveyFromJSON(filename string) OldSurvey {
	content, err := os.ReadFile(filename)
	if err != nil {
		logger.Error.Fatalf("Failed to read test-file: %s - %v", filename, err)
	}
	var oldSurveyDef OldSurvey
	err = json.Unmarshal(content, &oldSurveyDef)
	if err != nil {
		logger.Error.Fatal(err)
	}
	return oldSurveyDef
}

func saveNewSurveyIntoJSON(newSurveyHistory []*types.Survey, filename string) {
	newSurveyHistoryJSON := types.SurveyVersionsJSON{
		SurveyVersions: newSurveyHistory,
	}
	file, _ := json.Marshal(newSurveyHistoryJSON)

	err := os.WriteFile(filename, file, 0644)
	if err != nil {
		logger.Error.Fatal(err)
	}
}

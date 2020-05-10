package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/influenzanet/study-service/pkg/dbs/globaldb"
	"github.com/influenzanet/study-service/pkg/dbs/studydb"
	"github.com/influenzanet/study-service/pkg/types"

	"google.golang.org/grpc/status"
)

var testGlobalDBService *globaldb.GlobalDBService
var testStudyDBService *studydb.StudyDBService

const (
	testDBNamePrefix = "TEST_SERVICE_"
)

var (
	testInstanceID = strconv.FormatInt(time.Now().Unix(), 10)
)

// Pre-Test Setup
func TestMain(m *testing.M) {
	testGlobalDBService = setupTestGlobalDBService()
	testStudyDBService = setupTestStudyDBService()

	if testGlobalDBService == nil || testStudyDBService == nil {
		log.Fatal("no db services")
	}
	result := m.Run()
	dropTestDB()
	os.Exit(result)
}

func setupTestGlobalDBService() *globaldb.GlobalDBService {
	connStr := os.Getenv("GLOBAL_DB_CONNECTION_STR")
	username := os.Getenv("GLOBAL_DB_USERNAME")
	password := os.Getenv("GLOBAL_DB_PASSWORD")
	prefix := os.Getenv("GLOBAL_DB_CONNECTION_PREFIX") // Used in test mode
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
	return globaldb.NewGlobalDBService(
		types.DBConfig{
			URI:             URI,
			Timeout:         Timeout,
			IdleConnTimeout: IdleConnTimeout,
			MaxPoolSize:     MaxPoolSize,
			DBNamePrefix:    testDBNamePrefix,
		},
	)
}

func setupTestStudyDBService() *studydb.StudyDBService {
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
	return studydb.NewStudyDBService(
		types.DBConfig{
			URI:             URI,
			Timeout:         Timeout,
			IdleConnTimeout: IdleConnTimeout,
			MaxPoolSize:     MaxPoolSize,
			DBNamePrefix:    testDBNamePrefix,
		},
	)
}

func dropTestDB() {
	log.Println("Drop test database: service package")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := testStudyDBService.DBClient.Database(testDBNamePrefix + testInstanceID + "_studyDB").Drop(ctx)
	if err != nil {
		log.Fatal(err)
	}
	err = testGlobalDBService.DBClient.Database(testDBNamePrefix + "global-infos").Drop(ctx)
	if err != nil {
		log.Fatal(err)
	}
}

func shouldHaveGrpcErrorStatus(err error, expectedError string) (bool, string) {
	if err == nil {
		return false, "should return an error"
	}
	st, ok := status.FromError(err)
	if !ok || st == nil {
		return false, fmt.Sprintf("unexpected error: %s", err.Error())
	}

	if len(expectedError) > 0 && st.Message() != expectedError {
		return false, fmt.Sprintf("wrong error: %s", st.Message())
	}
	return true, ""
}

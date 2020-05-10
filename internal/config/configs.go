package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/influenzanet/study-service/pkg/types"
)

// Config is the structure that holds all global configuration data
type Config struct {
	Port           string
	StudyDBConfig  types.DBConfig
	GlobalDBConfig types.DBConfig
	Study          types.StudyConfig
}

func InitConfig() Config {
	conf := Config{}
	conf.Port = os.Getenv("STUDY_SERVICE_LISTEN_PORT")
	conf.StudyDBConfig = getStudyDBConfig()
	conf.GlobalDBConfig = getGlobalDBConfig()
	conf.Study = getStudyConfig()
	return conf
}

func getStudyConfig() types.StudyConfig {
	studyConf := types.StudyConfig{}
	studyConf.GlobalSecret = os.Getenv("STUDY_GLOBAL_SECRET")

	freq, err := strconv.Atoi(os.Getenv("STUDY_TIMER_EVENT_FREQUENCY"))
	if err != nil {
		log.Fatal("STUDY_TIMER_EVENT_FREQUENCY: " + err.Error())
	}
	studyConf.TimerEventFrequency = int64(freq)

	studyConf.TimerEventCheckIntervalMin, err = strconv.Atoi(os.Getenv("STUDY_TIMER_EVENT_CHECK_INTERVAL_MIN"))
	if err != nil {
		log.Fatal("STUDY_TIMER_EVENT_CHECK_INTERVAL_MIN: " + err.Error())
	}

	studyConf.TimerEventCheckIntervalVar, err = strconv.Atoi(os.Getenv("STUDY_TIMER_EVENT_CHECK_INTERVAL_VAR"))
	if err != nil {
		log.Fatal("STUDY_TIMER_EVENT_CHECK_INTERVAL: " + err.Error())
	}
	return studyConf
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

func getGlobalDBConfig() types.DBConfig {
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

	DBNamePrefix := os.Getenv("DB_DB_NAME_PREFIX")

	return types.DBConfig{
		URI:             URI,
		Timeout:         Timeout,
		IdleConnTimeout: IdleConnTimeout,
		MaxPoolSize:     MaxPoolSize,
		DBNamePrefix:    DBNamePrefix,
	}
}

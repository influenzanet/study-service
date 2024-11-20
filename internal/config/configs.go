package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/coneno/logger"
	"github.com/influenzanet/study-service/pkg/types"
	"gopkg.in/yaml.v2"
)

// Config is the structure that holds all global configuration data
type Config struct {
	Port           string
	LogLevel       logger.LogLevel
	StudyDBConfig  types.DBConfig
	GlobalDBConfig types.DBConfig
	Study          types.StudyConfig
	MaxMsgSize     int
	ServiceURLs    struct {
		LoggingService string
	}
	PersistentStoreConfig types.PersistentStoreConfig
	ExternalServices      []types.ExternalService
	DisableTimerTask      bool
}

func InitConfig() Config {
	conf := Config{}
	conf.Port = os.Getenv("STUDY_SERVICE_LISTEN_PORT")

	conf.MaxMsgSize = defaultGRPCMaxMsgSize
	ms, err := strconv.Atoi(os.Getenv(ENV_GRPC_MAX_MSG_SIZE))
	if err != nil {
		logger.Info.Printf("using default max gRPC message size: %d", defaultGRPCMaxMsgSize)
	} else {
		conf.MaxMsgSize = ms
	}

	conf.ServiceURLs.LoggingService = os.Getenv("ADDR_LOGGING_SERVICE")
	conf.LogLevel = getLogLevel()
	conf.StudyDBConfig = getStudyDBConfig()
	conf.GlobalDBConfig = getGlobalDBConfig()
	conf.Study = getStudyConfig()

	conf.PersistentStoreConfig = getPersistentStoreConfig()

	conf.ExternalServices = getExternalServicesConfig()

	conf.DisableTimerTask = os.Getenv("DISABLE_TIMER_TASK") == "true"
	return conf
}

func getLogLevel() logger.LogLevel {
	switch os.Getenv("LOG_LEVEL") {
	case "debug":
		return logger.LEVEL_DEBUG
	case "info":
		return logger.LEVEL_INFO
	case "error":
		return logger.LEVEL_ERROR
	case "warning":
		return logger.LEVEL_WARNING
	default:
		return logger.LEVEL_INFO
	}
}

func getPersistentStoreConfig() types.PersistentStoreConfig {
	c := types.PersistentStoreConfig{}
	c.RootPath = os.Getenv(ENV_PERSISTENCE_STORE_ROOT_PATH)
	if len(c.RootPath) < 1 {
		c.RootPath = defaultPersistenceStoreRootPath
	}

	val, err := strconv.Atoi(os.Getenv(ENV_PERSISTENCE_MAX_FILE_SIZE))
	if err != nil || val < 1 {
		c.MaxParticipantFileSize = maxParticipantFileSize
		logger.Info.Printf("Using default value for participant file max size: %d bytes. Reason: couldn't read env variable: %v", c.MaxParticipantFileSize, err.Error())
	} else {
		c.MaxParticipantFileSize = val
		logger.Info.Printf("Participant file max size: %d bytes", c.MaxParticipantFileSize)
	}

	return c
}

func getStudyConfig() types.StudyConfig {
	studyConf := types.StudyConfig{}
	studyConf.GlobalSecret = os.Getenv("STUDY_GLOBAL_SECRET")

	freq, err := strconv.Atoi(os.Getenv("STUDY_TIMER_EVENT_FREQUENCY"))
	if err != nil {
		logger.Error.Fatal("STUDY_TIMER_EVENT_FREQUENCY: " + err.Error())
	}
	studyConf.TimerEventFrequency = int64(freq)

	studyConf.TimerEventCheckIntervalMin, err = strconv.Atoi(os.Getenv("STUDY_TIMER_EVENT_CHECK_INTERVAL_MIN"))
	if err != nil {
		logger.Error.Fatal("STUDY_TIMER_EVENT_CHECK_INTERVAL_MIN: " + err.Error())
	}

	studyConf.TimerEventCheckIntervalVar, err = strconv.Atoi(os.Getenv("STUDY_TIMER_EVENT_CHECK_INTERVAL_VAR"))
	if err != nil {
		logger.Error.Fatal("STUDY_TIMER_EVENT_CHECK_INTERVAL: " + err.Error())
	}
	return studyConf
}

func getStudyDBConfig() types.DBConfig {
	connStr := os.Getenv("STUDY_DB_CONNECTION_STR")
	username := os.Getenv("STUDY_DB_USERNAME")
	password := os.Getenv("STUDY_DB_PASSWORD")
	prefix := os.Getenv("STUDY_DB_CONNECTION_PREFIX") // Used in test mode
	if connStr == "" || username == "" || password == "" {
		logger.Error.Fatal("Couldn't read DB credentials.")
	}
	URI := fmt.Sprintf(`mongodb%s://%s:%s@%s`, prefix, username, password, connStr)

	var err error
	Timeout, err := strconv.Atoi(os.Getenv("DB_TIMEOUT"))
	if err != nil {
		logger.Error.Fatal("DB_TIMEOUT: " + err.Error())
	}
	IdleConnTimeout, err := strconv.Atoi(os.Getenv("DB_IDLE_CONN_TIMEOUT"))
	if err != nil {
		logger.Error.Fatal("DB_IDLE_CONN_TIMEOUT" + err.Error())
	}
	mps, err := strconv.Atoi(os.Getenv("DB_MAX_POOL_SIZE"))
	MaxPoolSize := uint64(mps)
	if err != nil {
		logger.Error.Fatal("DB_MAX_POOL_SIZE: " + err.Error())
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
		logger.Error.Fatal("Couldn't read DB credentials.")
	}
	URI := fmt.Sprintf(`mongodb%s://%s:%s@%s`, prefix, username, password, connStr)

	var err error
	Timeout, err := strconv.Atoi(os.Getenv("DB_TIMEOUT"))
	if err != nil {
		logger.Error.Fatal("DB_TIMEOUT: " + err.Error())
	}
	IdleConnTimeout, err := strconv.Atoi(os.Getenv("DB_IDLE_CONN_TIMEOUT"))
	if err != nil {
		logger.Error.Fatal("DB_IDLE_CONN_TIMEOUT" + err.Error())
	}
	mps, err := strconv.Atoi(os.Getenv("DB_MAX_POOL_SIZE"))
	MaxPoolSize := uint64(mps)
	if err != nil {
		logger.Error.Fatal("DB_MAX_POOL_SIZE: " + err.Error())
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

func getExternalServicesConfig() []types.ExternalService {
	configFilePath := os.Getenv(ENV_EXTERNAL_SERVICES_CONFIG_PATH)

	if configFilePath == "" {
		return []types.ExternalService{}
	}

	yamlFile, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		logger.Error.Fatalf("cannot read external services config file: %v ", err)
		return []types.ExternalService{}
	}

	var services types.ExternalServices
	err = yaml.UnmarshalStrict(yamlFile, &services)
	if err != nil {
		logger.Error.Fatalf("cannot parse external services config file: %v ", err)
		return []types.ExternalService{}
	}
	return services.Services
}

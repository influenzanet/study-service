package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

// Config is the structure that holds all global configuration data
type Config struct {
	Port string
	DB   struct {
		URI             string
		DBNamePrefix    string
		Timeout         int
		MaxPoolSize     uint64
		IdleConnTimeout int
	}
	Study struct {
		GlobalSecret               string
		TimerEventFrequency        int64 // how often the timer event should be performed (only from one instance of the service) - seconds
		TimerEventCheckIntervalMin int   // approx. how often this serice should check if to perform the timer event - seconds
		TimerEventCheckIntervalVar int   // range of the uniform random distribution - varying the check interval to avoid a steady collisions
	}
}

func initConfig() {
	conf.Port = os.Getenv("STUDY_SERVICE_LISTEN_PORT")
	getStudyConfig()
	getDBConfig()
}

func getStudyConfig() {
	conf.Study.GlobalSecret = os.Getenv("STUDY_GLOBAL_SECRET")

	freq, err := strconv.Atoi(os.Getenv("STUDY_TIMER_EVENT_FREQUENCY"))
	if err != nil {
		log.Fatal("STUDY_TIMER_EVENT_FREQUENCY: " + err.Error())
	}
	conf.Study.TimerEventFrequency = int64(freq)

	conf.Study.TimerEventCheckIntervalMin, err = strconv.Atoi(os.Getenv("STUDY_TIMER_EVENT_CHECK_INTERVAL_MIN"))
	if err != nil {
		log.Fatal("STUDY_TIMER_EVENT_CHECK_INTERVAL_MIN: " + err.Error())
	}

	conf.Study.TimerEventCheckIntervalVar, err = strconv.Atoi(os.Getenv("STUDY_TIMER_EVENT_CHECK_INTERVAL_VAR"))
	if err != nil {
		log.Fatal("STUDY_TIMER_EVENT_CHECK_INTERVAL: " + err.Error())
	}
}

func getDBConfig() {
	connStr := os.Getenv("DB_CONNECTION_STR")
	username := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	prefix := os.Getenv("DB_PREFIX") // Used in test mode
	if connStr == "" || username == "" || password == "" {
		log.Fatal("Couldn't read DB credentials.")
	}
	conf.DB.URI = fmt.Sprintf(`mongodb%s://%s:%s@%s`, prefix, username, password, connStr)

	var err error
	conf.DB.Timeout, err = strconv.Atoi(os.Getenv("DB_TIMEOUT"))
	if err != nil {
		log.Fatal("DB_TIMEOUT: " + err.Error())
	}
	conf.DB.IdleConnTimeout, err = strconv.Atoi(os.Getenv("DB_IDLE_CONN_TIMEOUT"))
	if err != nil {
		log.Fatal("DB_IDLE_CONN_TIMEOUT" + err.Error())
	}
	mps, err := strconv.Atoi(os.Getenv("DB_MAX_POOL_SIZE"))
	conf.DB.MaxPoolSize = uint64(mps)
	if err != nil {
		log.Fatal("DB_MAX_POOL_SIZE: " + err.Error())
	}

	conf.DB.DBNamePrefix = os.Getenv("DB_DB_NAME_PREFIX")
}

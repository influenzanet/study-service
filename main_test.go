package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"testing"
	"time"
)

var testInstanceID = strconv.FormatInt(time.Now().Unix(), 10)

// Pre-Test Setup
func TestMain(m *testing.M) {
	result := m.Run()
	dropTestDB()
	os.Exit(result)
}

func dropTestDB() {
	log.Println("Drop test database")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := dbClient.Database(conf.DB.DBNamePrefix + testInstanceID + "_studies").Drop(ctx)
	if err != nil {
		log.Fatal(err)
	}
}

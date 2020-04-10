package main

import (
	"context"
	"log"
	"time"

	"github.com/influenzanet/study-service/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func getAllInstances() ([]models.Instance, error) {
	coll := dbClient.Database(conf.DB.DBNamePrefix + "global-infos").Collection("instances")
	ctx, cancel := getContext()
	defer cancel()

	filter := bson.M{}
	cur, err := coll.Find(
		ctx,
		filter,
	)

	if err != nil {
		return []models.Instance{}, err
	}
	defer cur.Close(ctx)

	instances := []models.Instance{}
	for cur.Next(ctx) {
		var result models.Instance
		err := cur.Decode(&result)
		if err != nil {
			return instances, err
		}

		instances = append(instances, result)
	}
	if err := cur.Err(); err != nil {
		return instances, err
	}

	return instances, nil
}

// Collections
func collectionRefStudyInfos(instanceID string) *mongo.Collection {
	return dbClient.Database(conf.DB.DBNamePrefix + instanceID + "_studyDB").Collection("study-infos")
}

func collectionRefStudyParticipant(instanceID string, studyKey string) *mongo.Collection {
	return dbClient.Database(conf.DB.DBNamePrefix + instanceID + "_studyDB").Collection(studyKey + "_participants")
}

func collectionRefStudySurveys(instanceID string, studyKey string) *mongo.Collection {
	return dbClient.Database(conf.DB.DBNamePrefix + instanceID + "_studyDB").Collection(studyKey + "_surveys")
}

func collectionRefStudyMessageBoard(instanceID string, studyKey string) *mongo.Collection {
	return dbClient.Database(conf.DB.DBNamePrefix + instanceID + "_studyDB").Collection(studyKey + "_messageBoard")
}

func collectionRefSurveyResponses(instanceID string, studyKey string) *mongo.Collection {
	return dbClient.Database(conf.DB.DBNamePrefix + instanceID + "_studyDB").Collection(studyKey + "_surveyResponses")
}

// DB utils
func getContext() (ctx context.Context, cancel context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Duration(conf.DB.Timeout)*time.Second)
}

// Connect to DB
func dbInit() {
	var err error
	dbClient, err = mongo.NewClient(
		options.Client().ApplyURI(conf.DB.URI),
		options.Client().SetMaxConnIdleTime(time.Duration(conf.DB.IdleConnTimeout)*time.Second),
		options.Client().SetMaxPoolSize(conf.DB.MaxPoolSize),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.DB.Timeout)*time.Second)
	defer cancel()

	err = dbClient.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	ctx, conCancel := context.WithTimeout(context.Background(), time.Duration(conf.DB.Timeout)*time.Second)
	err = dbClient.Ping(ctx, nil)
	defer conCancel()
	if err != nil {
		log.Fatal("fail to connect to DB: " + err.Error())
	}
}

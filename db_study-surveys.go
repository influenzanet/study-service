package main

import (
	"github.com/influenzanet/study-service/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func addSurveyToDB(instanceID string, studyKey string, survey models.Survey) (newSurvey models.Survey, err error) {
	ctx, cancel := getContext()
	defer cancel()

	res, err := collectionRefStudySurveys(instanceID, studyKey).InsertOne(ctx, survey)
	if err != nil {
		return
	}

	newSurvey = survey
	newSurvey.ID = res.InsertedID.(primitive.ObjectID)
	return
}

func findSurveyDefDB(instanceID string, studyKey string, surveyKey string) (models.Survey, error) {
	ctx, cancel := getContext()
	defer cancel()

	filter := bson.M{"current.surveyDefinition.key": surveyKey}

	elem := models.Survey{}
	err := collectionRefStudySurveys(instanceID, studyKey).FindOne(ctx, filter).Decode(&elem)
	return elem, err
}

func findAllSurveyDefsForStudyDB(instanceID string, studyKey string, onlyInfos bool) (surveys []models.Survey, err error) {
	ctx, cancel := getContext()
	defer cancel()

	filter := bson.M{}

	var opts *options.FindOptions
	if onlyInfos {
		projection := bson.D{
			primitive.E{Key: "key", Value: 1},
			primitive.E{Key: "name", Value: 1},
			primitive.E{Key: "description", Value: 1},
		}
		opts = options.Find().SetProjection(projection)
	}

	cur, err := collectionRefStudySurveys(instanceID, studyKey).Find(
		ctx,
		filter,
		opts,
	)

	if err != nil {
		return surveys, err
	}
	defer cur.Close(ctx)

	surveys = []models.Survey{}
	for cur.Next(ctx) {
		var result models.Survey
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

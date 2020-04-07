package main

import (
	"github.com/influenzanet/study-service/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

package main

import (
	"github.com/influenzanet/study-service/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func addSurveyToDB(instanceID string, survey models.Survey) (newSurvey models.Survey, err error) {
	ctx, cancel := getContext()
	defer cancel()

	res, err := collectionRefSurveys(instanceID).InsertOne(ctx, survey)
	if err != nil {
		return
	}

	newSurvey = survey
	newSurvey.ID = res.InsertedID.(primitive.ObjectID)
	return
}

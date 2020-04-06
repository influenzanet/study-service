package main

import (
	"github.com/influenzanet/study-service/models"
)

func addSurveyResponseToDB(instanceID string, studyKey string, response models.SurveyResponse) error {
	ctx, cancel := getContext()
	defer cancel()

	_, err := collectionRefSurveyResponses(instanceID, studyKey).InsertOne(ctx, response)
	return err
}

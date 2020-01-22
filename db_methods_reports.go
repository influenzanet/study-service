package main

import (
	"log"

	"github.com/influenzanet/study-service/models"
)

func addSurveyResponseToDB(instanceID string, studyID string, report models.SurveyResponseReport) error {
	ctx, cancel := getContext()
	defer cancel()

	_, err := collectionRefReports(instanceID, studyID).InsertOne(ctx, report)
	if err != nil {
		return err
	}
	log.Printf("new report submitted for study %s", studyID)
	return nil
}

package main

import (
	"errors"

	"github.com/influenzanet/study-service/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func addSurveyResponseToDB(instanceID string, studyKey string, response models.SurveyResponse) error {
	ctx, cancel := getContext()
	defer cancel()

	_, err := collectionRefSurveyResponses(instanceID, studyKey).InsertOne(ctx, response)
	return err
}

type responseQuery struct {
	ParticipantID string
	SurveyKey     string
	Limit         int64
	Since         int64
}

func findSurveyResponsesInDB(instanceID string, studyKey string, query responseQuery) (responses []models.SurveyResponse, err error) {
	ctx, cancel := getContext()
	defer cancel()

	if query.ParticipantID == "" {
		return responses, errors.New("participant id must be defined")
	}
	filter := bson.M{"participantID": query.ParticipantID}

	opts := &options.FindOptions{
		Sort: bson.D{
			primitive.E{Key: "submittedAt", Value: -1},
		},
	}

	if query.Limit > 0 {
		opts.SetLimit(query.Limit)
	}

	if len(query.SurveyKey) > 0 {
		filter["key"] = query.SurveyKey
	}

	if query.Since > 0 {
		filter["submittedAt"] = bson.M{"$gt": query.Since}
	}

	cur, err := collectionRefSurveyResponses(instanceID, studyKey).Find(
		ctx,
		filter,
		opts,
	)

	if err != nil {
		return responses, err
	}
	defer cur.Close(ctx)

	responses = []models.SurveyResponse{}
	for cur.Next(ctx) {
		var result models.SurveyResponse
		err := cur.Decode(&result)
		if err != nil {
			return responses, err
		}

		responses = append(responses, result)
	}
	if err := cur.Err(); err != nil {
		return responses, err
	}

	return responses, nil
}

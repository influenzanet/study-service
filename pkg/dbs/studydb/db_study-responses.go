package studydb

import (
	"errors"
	"time"

	"github.com/influenzanet/study-service/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (dbService *StudyDBService) AddSurveyResponse(instanceID string, studyKey string, response types.SurveyResponse) error {
	ctx, cancel := dbService.getContext()
	defer cancel()
	response.ArrivedAt = time.Now().Unix()
	_, err := dbService.collectionRefSurveyResponses(instanceID, studyKey).InsertOne(ctx, response)
	return err
}

type ResponseQuery struct {
	ParticipantID string
	SurveyKey     string
	Limit         int64
	Since         int64
}

func (dbService *StudyDBService) FindSurveyResponses(instanceID string, studyKey string, query ResponseQuery) (responses []types.SurveyResponse, err error) {
	ctx, cancel := dbService.getContext()
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

	cur, err := dbService.collectionRefSurveyResponses(instanceID, studyKey).Find(
		ctx,
		filter,
		opts,
	)

	if err != nil {
		return responses, err
	}
	defer cur.Close(ctx)

	responses = []types.SurveyResponse{}
	for cur.Next(ctx) {
		var result types.SurveyResponse
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

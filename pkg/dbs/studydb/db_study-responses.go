package studydb

import (
	"errors"
	"fmt"
	"log"
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

func (dbService *StudyDBService) GetSurveyResponseKeys(instanceID string, studyKey string, from int64, until int64) (keys []string, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{}

	if from > 0 && until > 0 {
		filter["$and"] = bson.A{
			bson.M{"submittedAt": bson.M{"$gt": from}},
			bson.M{"submittedAt": bson.M{"$lt": until}},
		}
	} else if from > 0 {
		filter["submittedAt"] = bson.M{"$gt": from}
	} else if until > 0 {
		filter["submittedAt"] = bson.M{"$lt": until}
	}

	k, err := dbService.collectionRefSurveyResponses(instanceID, studyKey).Distinct(ctx, "key", filter)
	if err != nil {
		return keys, err
	}
	keys = make([]string, len(k))
	for i, v := range k {
		keys[i] = fmt.Sprint(v)
	}
	return keys, err
}

func (dbService *StudyDBService) CountSurveyResponsesByKey(instanceID string, studyKey string, surveyKey string, from int64, until int64) (count int64, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{}
	if len(surveyKey) > 0 {
		filter["key"] = surveyKey
	}
	if from > 0 && until > 0 {
		filter["$and"] = bson.A{
			bson.M{"submittedAt": bson.M{"$gt": from}},
			bson.M{"submittedAt": bson.M{"$lt": until}},
		}
	} else if from > 0 {
		filter["submittedAt"] = bson.M{"$gt": from}
	} else if until > 0 {
		filter["submittedAt"] = bson.M{"$lt": until}
	}

	count, err = dbService.collectionRefSurveyResponses(instanceID, studyKey).CountDocuments(ctx, filter)
	return count, err
}

func (dbService *StudyDBService) PerfomActionForSurveyResponses(
	instanceID string,
	studyKey string, surveyKey string, from int64, until int64,
	cbk func(instanceID string, studyKey string, response types.SurveyResponse, args ...interface{}) error,
	args ...interface{},
) (err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{}
	if len(surveyKey) > 0 {
		filter["key"] = surveyKey
	}
	if from > 0 && until > 0 {
		filter["$and"] = bson.A{
			bson.M{"submittedAt": bson.M{"$gt": from}},
			bson.M{"submittedAt": bson.M{"$lt": until}},
		}
	} else if from > 0 {
		filter["submittedAt"] = bson.M{"$gt": from}
	} else if until > 0 {
		filter["submittedAt"] = bson.M{"$lt": until}
	}

	cur, err := dbService.collectionRefSurveyResponses(instanceID, studyKey).Find(
		ctx,
		filter,
	)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var result types.SurveyResponse
		err := cur.Decode(&result)
		if err != nil {
			return err
		}

		if err := cbk(instanceID, studyKey, result, args...); err != nil {
			log.Printf("PerfomActionForSurveyResponses: %v", err)
		}
	}
	if err := cur.Err(); err != nil {
		return err
	}
	return nil
}

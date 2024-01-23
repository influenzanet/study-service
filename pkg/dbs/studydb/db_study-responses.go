package studydb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/coneno/logger"
	"github.com/influenzanet/study-service/pkg/types"
	"github.com/influenzanet/study-service/pkg/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (dbService *StudyDBService) AddSurveyResponse(instanceID string, studyKey string, response types.SurveyResponse) (string, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()
	response.ArrivedAt = time.Now().Unix()
	res, err := dbService.collectionRefSurveyResponses(instanceID, studyKey).InsertOne(ctx, response)
	id := res.InsertedID.(primitive.ObjectID)
	return id.Hex(), err
}

type ResponseQuery struct {
	ParticipantID string
	SurveyKey     string
	Limit         int64
	Since         int64
	Until         int64
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

	if query.Since > 0 && query.Until > 0 {
		filter["$and"] = bson.A{
			bson.M{"submittedAt": bson.M{"$gt": query.Since}},
			bson.M{"submittedAt": bson.M{"$lt": query.Until}},
		}
	} else if query.Since > 0 {
		filter["submittedAt"] = bson.M{"$gt": query.Since}
	} else if query.Until > 0 {
		filter["submittedAt"] = bson.M{"$lt": query.Until}
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

func (dbService *StudyDBService) GetSurveyResponsesCount(
	ctx context.Context,
	instanceID string,
	studyKey string, surveyKey string, from int64, until int64) (totalCount int32) {
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
	count, err := dbService.collectionRefSurveyResponses(instanceID, studyKey).CountDocuments(
		ctx,
		filter,
	)
	totalCount = int32(count)
	if err != nil {
		return 0
	} else {
		return totalCount
	}
}

func (dbService *StudyDBService) PerformActionForSurveyResponses(
	ctx context.Context,
	instanceID string,
	studyKey string, surveyKey string, from int64, until int64,
	cbk func(instanceID string, studyKey string, response types.SurveyResponse, args ...interface{}) error,
	args ...interface{},
) (err error) {
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
	count, err := dbService.collectionRefSurveyResponses(instanceID, studyKey).CountDocuments(
		ctx,
		filter,
	)
	totalCount := int32(count)
	if err != nil {
		return err
	}

	batchSize := int32(32)
	opts := options.FindOptions{
		BatchSize: &batchSize,
	}
	page := int32(0)
	pageSize := int32(0)
	for index, val := range args {
		switch index {
		case 1: //page is optional param
			page, _ = val.(int32)
		case 2: //pageSize is optional param
			pageSize, _ = val.(int32)
		}
	}
	if utils.CheckForValidPaginationParameter(pageSize, page) {
		pageCount := utils.ComputePageCount(pageSize, totalCount)
		if page > pageCount {
			if pageCount > 0 {
				page = pageCount
			} else {
				page = 1
			}
		}
		opts.SetSkip((int64(page) - 1) * int64(pageSize))
		opts.SetLimit(int64(pageSize))
	}

	cur, err := dbService.collectionRefSurveyResponses(instanceID, studyKey).Find(
		ctx,
		filter,
		&opts,
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
			logger.Error.Println(err)
		}
	}
	if err := cur.Err(); err != nil {
		return err
	}
	return nil
}

func (dbService *StudyDBService) UpdateParticipantIDonResponses(instanceID string, studyKey string, oldID string, newID string) (count int64, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	if oldID == "" || newID == "" {
		return 0, errors.New("participant id must be defined")
	}
	filter := bson.M{"participantID": oldID}
	update := bson.M{"$set": bson.M{"participantID": newID}}

	res, err := dbService.collectionRefSurveyResponses(instanceID, studyKey).UpdateMany(ctx, filter, update)
	return res.ModifiedCount, err
}

func (dbService *StudyDBService) DeleteSurveyResponses(instanceID string, studyKey string, query ResponseQuery) (count int64, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	if query.ParticipantID == "" {
		return 0, errors.New("participant id must be defined")
	}
	filter := bson.M{"participantID": query.ParticipantID}
	if query.SurveyKey != "" {
		filter["key"] = query.SurveyKey
	}

	res, err := dbService.collectionRefSurveyResponses(instanceID, studyKey).DeleteMany(ctx, filter)
	return res.DeletedCount, err
}

func (dbService *StudyDBService) CreateParticipantIDIndexForResponses(instanceID string, studyKey string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_, err := dbService.collectionRefSurveyResponses(instanceID, studyKey).Indexes().CreateOne(
		ctx, mongo.IndexModel{
			Keys: bson.D{
				{Key: "participantID", Value: 1},
			},
		},
	)
	return err
}

func (dbService *StudyDBService) CreateParticipantIDIndexForResponsesForAllStudies(instanceID string) {
	studies, err := dbService.GetStudiesByStatus(instanceID, "", true)
	if err != nil {
		logger.Error.Printf("unexpected error when fetching studies in '%s': %v", instanceID, err)
		return
	}

	for _, study := range studies {
		err = dbService.CreateParticipantIDIndexForResponses(instanceID, study.Key)
		if err != nil {
			logger.Error.Printf("unexpected error when creating participantID index for survey responses for study: %v, %v", err, study)
		}
	}
}

func (dbService *StudyDBService) CreateSubmittedAtIndex(instanceID string, studyKey string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_, err := dbService.collectionRefSurveyResponses(instanceID, studyKey).Indexes().CreateOne(
		ctx, mongo.IndexModel{
			Keys: bson.D{
				{Key: "submittedAt", Value: 1},
			},
		},
	)
	return err
}

func (dbService *StudyDBService) CreateSubmittedAtIndexForAllStudies(instanceID string) {
	studies, err := dbService.GetStudiesByStatus(instanceID, "", true)
	if err != nil {
		logger.Error.Printf("unexpected error when fetching studies in '%s': %v", instanceID, err)
		return
	}

	for _, study := range studies {
		err = dbService.CreateSubmittedAtIndex(instanceID, study.Key)
		if err != nil {
			logger.Error.Printf("unexpected error when creating submittedAt index for survey responses for study: %v, %v", err, study)
		}
	}
}

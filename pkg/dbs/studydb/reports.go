package studydb

import (
	"errors"

	"github.com/influenzanet/study-service/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (dbService *StudyDBService) SaveReport(instanceID string, studyKey string, report types.Report) error {
	ctx, cancel := dbService.getContext()
	defer cancel()
	_, err := dbService.collectionRefReportHistory(instanceID, studyKey).InsertOne(ctx, report)
	return err
}

type ReportQuery struct {
	ParticipantID string
	Key           string
	Limit         int64
	Since         int64
	Until         int64
}

func (dbService *StudyDBService) FindReports(instanceID string, studyKey string, query ReportQuery) (responses []types.Report, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	if query.ParticipantID == "" {
		return responses, errors.New("participant id must be defined")
	}
	filter := bson.M{"participantID": query.ParticipantID}

	opts := &options.FindOptions{
		Sort: bson.D{
			primitive.E{Key: "timestamp", Value: -1},
		},
	}

	if query.Limit > 0 {
		opts.SetLimit(query.Limit)
	}

	if len(query.Key) > 0 {
		filter["key"] = query.Key
	}

	if query.Since > 0 && query.Until > 0 {
		filter["$and"] = bson.A{
			bson.M{"timestamp": bson.M{"$gt": query.Since}},
			bson.M{"timestamp": bson.M{"$lt": query.Until}},
		}
	} else if query.Since > 0 {
		filter["timestamp"] = bson.M{"$gt": query.Since}
	} else if query.Until > 0 {
		filter["timestamp"] = bson.M{"$lt": query.Until}
	}

	cur, err := dbService.collectionRefReportHistory(instanceID, studyKey).Find(
		ctx,
		filter,
		opts,
	)

	if err != nil {
		return responses, err
	}
	defer cur.Close(ctx)

	responses = []types.Report{}
	for cur.Next(ctx) {
		var result types.Report
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

func (dbService *StudyDBService) UpdateParticipantIDonReports(instanceID string, studyKey string, oldID string, newID string) (count int64, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	if oldID == "" || newID == "" {
		return 0, errors.New("participant id must be defined")
	}
	filter := bson.M{"participantID": oldID}
	update := bson.M{"$set": bson.M{"participantID": newID}}

	res, err := dbService.collectionRefReportHistory(instanceID, studyKey).UpdateMany(ctx, filter, update)
	return res.ModifiedCount, err
}

package studydb

import (
	"context"
	"errors"

	"github.com/coneno/logger"
	"github.com/influenzanet/study-service/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (dbService *StudyDBService) SaveFileInfo(instanceID string, studyKey string, fileInfo types.FileInfo) (types.FileInfo, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	if fileInfo.ID.IsZero() {
		fileInfo.ID = primitive.NewObjectID()
	}

	filter := bson.M{"_id": fileInfo.ID}

	upsert := true
	rd := options.After
	options := options.FindOneAndReplaceOptions{
		Upsert:         &upsert,
		ReturnDocument: &rd,
	}
	elem := types.FileInfo{}
	err := dbService.collectionRefParticipantFiles(instanceID, studyKey).FindOneAndReplace(
		ctx, filter, fileInfo, &options,
	).Decode(&elem)
	return elem, err
}

func (dbService *StudyDBService) FindFileInfo(instanceID string, studyKey string, fileID string) (types.FileInfo, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_id, _ := primitive.ObjectIDFromHex(fileID)
	filter := bson.M{"_id": _id}

	elem := types.FileInfo{}
	err := dbService.collectionRefParticipantFiles(instanceID, studyKey).FindOne(ctx, filter).Decode(&elem)
	return elem, err
}

type FileInfoQuery struct {
	ParticipantID string
	FileType      string
	Since         int64
	Until         int64
}

func (dbService *StudyDBService) PerformActionForFileInfos(
	ctx context.Context,
	instanceID string,
	studyKey string,
	query FileInfoQuery,
	cbk func(instanceID string, studyKey string, fileInfo types.FileInfo, args ...interface{}) error,
	args ...interface{},
) (err error) {
	filter := bson.M{}
	if len(query.ParticipantID) > 0 {
		filter["participantID"] = query.ParticipantID
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

	cur, err := dbService.collectionRefParticipantFiles(instanceID, studyKey).Find(
		ctx,
		filter,
	)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var result types.FileInfo
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

func (dbService *StudyDBService) DeleteFileInfo(instanceID string, studyKey string, fileID string) (count int64, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	if fileID == "" {
		return 0, errors.New("file id must be defined")
	}
	_id, _ := primitive.ObjectIDFromHex(fileID)
	filter := bson.M{"_id": _id}

	res, err := dbService.collectionRefParticipantFiles(instanceID, studyKey).DeleteOne(ctx, filter)
	return res.DeletedCount, err
}

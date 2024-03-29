package studydb

import (
	"github.com/coneno/logger"
	"github.com/influenzanet/study-service/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (dbService *StudyDBService) SaveResearcherMessage(instanceID string, studyKey string, message types.StudyMessage) error {
	ctx, cancel := dbService.getContext()
	defer cancel()
	_, err := dbService.collectionRefResearcherMessages(instanceID, studyKey).InsertOne(ctx, message)
	return err
}

func (dbService *StudyDBService) FindResearcherMessages(instanceID string, studyKey string) (messages []types.StudyMessage, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{}

	cur, err := dbService.collectionRefResearcherMessages(instanceID, studyKey).Find(
		ctx,
		filter,
	)
	if err != nil {
		return messages, err
	}
	defer cur.Close(ctx)

	messages = []types.StudyMessage{}
	for cur.Next(ctx) {
		var result types.StudyMessage
		err := cur.Decode(&result)
		if err != nil {
			return messages, err
		}

		messages = append(messages, result)
	}
	if err := cur.Err(); err != nil {
		return messages, err
	}

	return messages, nil
}

func (dbService *StudyDBService) DeleteResearcherMessages(instanceID string, studyKey string, messageIDs []string) (int64, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	idsToDelete := []primitive.ObjectID{}
	for _, id := range messageIDs {
		_id, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			logger.Error.Printf("unexpected error: %v", err)
			continue
		}
		idsToDelete = append(idsToDelete, _id)
	}
	filter := bson.M{"_id": bson.M{"$in": idsToDelete}}

	res, err := dbService.collectionRefResearcherMessages(instanceID, studyKey).DeleteMany(ctx, filter)
	return res.DeletedCount, err
}

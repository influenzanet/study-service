package studydb

import (
	"context"
	"time"

	"github.com/coneno/logger"
	"github.com/influenzanet/study-service/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// findParticipantsByStudyStatusDB retrieves all participant states from a study by status (e.g. active)
func (dbService *StudyDBService) FindParticipantsByStudyStatus(instanceID string, studyKey string, studyStatus string, useProjection bool) (pStates []types.ParticipantState, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{"studyStatus": studyStatus}

	batchSize := int32(32)
	opts := options.FindOptions{
		BatchSize: &batchSize,
	}
	if useProjection {
		projection := bson.D{
			primitive.E{Key: "studyStatus", Value: 1},   // {"secretKey", 1},
			primitive.E{Key: "participantID", Value: 1}, // {"secretKey", 1},
		}
		opts.Projection = projection
	}

	cur, err := dbService.collectionRefStudyParticipant(instanceID, studyKey).Find(
		ctx,
		filter,
		&opts,
	)

	if err != nil {
		return pStates, err
	}
	defer cur.Close(ctx)

	pStates = []types.ParticipantState{}
	for cur.Next(ctx) {
		var result types.ParticipantState
		err := cur.Decode(&result)
		if err != nil {
			return pStates, err
		}

		pStates = append(pStates, result)
	}
	if err := cur.Err(); err != nil {
		return pStates, err
	}

	return pStates, nil
}

// findParticipantsByQuery retrieves participants that fulfill criteria of queryString with pagination... TODO: description here
func (dbService *StudyDBService) FindParticipantsByQuery(instanceID string, studyKey string, queryString string, limit int32, start int32) (pStates []types.ParticipantState, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{}
	err = bson.UnmarshalExtJSON([]byte(queryString), true, &filter)
	if err != nil {
		logger.Error.Println("Failed to parse query string:", err)
		return pStates, err
	}

	batchSize := int32(32)
	opts := options.FindOptions{
		BatchSize: &batchSize,
	}

	cur, err := dbService.collectionRefStudyParticipant(instanceID, studyKey).Find(
		ctx,
		filter,
		&opts,
	)

	if err != nil {
		return pStates, err
	}
	defer cur.Close(ctx)

	pStates = []types.ParticipantState{}
	for cur.Next(ctx) {
		var result types.ParticipantState
		err := cur.Decode(&result)
		if err != nil {
			return pStates, err
		}

		pStates = append(pStates, result)
	}
	if err := cur.Err(); err != nil {
		return pStates, err
	}

	return pStates, nil
}

// FindParticipantState retrieves the participant state for a given participant from a study
func (dbService *StudyDBService) FindParticipantState(instanceID string, studyKey string, participantID string) (types.ParticipantState, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{"participantID": participantID}

	elem := types.ParticipantState{}
	err := dbService.collectionRefStudyParticipant(instanceID, studyKey).FindOne(ctx, filter).Decode(&elem)
	return elem, err
}

// SaveParticipantState creates or replaces the participant states in the DB
func (dbService *StudyDBService) SaveParticipantState(instanceID string, studyKey string, pState types.ParticipantState) (types.ParticipantState, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{"participantID": pState.ParticipantID}

	upsert := true
	rd := options.After
	options := options.FindOneAndReplaceOptions{
		Upsert:         &upsert,
		ReturnDocument: &rd,
	}
	elem := types.ParticipantState{}
	err := dbService.collectionRefStudyParticipant(instanceID, studyKey).FindOneAndReplace(
		ctx, filter, pState, &options,
	).Decode(&elem)
	return elem, err
}

func (dbService *StudyDBService) DeleteParticipantState(instanceID string, studyKey string, pID string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{"participantID": pID}
	_, err := dbService.collectionRefStudyParticipant(instanceID, studyKey).DeleteOne(ctx, filter)
	return err
}

func (dbService *StudyDBService) GetParticipantCountByStatus(instanceID string, studyKey string, studyStatus string) (count int64, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{
		"studyStatus": studyStatus,
	}

	count, err = dbService.collectionRefStudyParticipant(instanceID, studyKey).CountDocuments(ctx, filter)
	return count, err
}

// FindParticipantState retrieves the participant state for a given participant from a study
func (dbService *StudyDBService) FindAndExecuteOnParticipantsStates(
	ctx context.Context,
	instanceID string,
	studyKey string,
	filterByStatus string,
	cbk func(dbService *StudyDBService, p types.ParticipantState, instanceID string, studyKey string, args ...interface{}) error,
	args ...interface{},
) error {
	filter := bson.M{}
	if len(filterByStatus) > 0 {
		filter["studyStatus"] = filterByStatus
	}

	batchSize := int32(32)
	options := options.FindOptions{
		BatchSize: &batchSize,
	}

	cur, err := dbService.collectionRefStudyParticipant(instanceID, studyKey).Find(ctx, filter, &options)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		if ctx.Err() != nil {
			logger.Debug.Println(ctx.Err())
			return ctx.Err()
		}
		// Update state of every participant
		var pState types.ParticipantState
		if err := cur.Decode(&pState); err != nil {
			logger.Error.Printf("wrong data model: %v, %v", pState, err)
			continue
		}
		// Perform callback:
		if err := cbk(dbService, pState, instanceID, studyKey, args...); err != nil {
			continue
		}
	}
	if err := cur.Err(); err != nil {
		return err
	}
	return nil
}

func (dbService *StudyDBService) DeleteMessagesFromParticipant(instanceID string, studyKey string, participantID string, messageIDs []string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{"participantID": participantID}
	update := bson.M{"$pull": bson.M{"messages": bson.M{
		"id": bson.M{"$in": messageIDs},
	}}}
	_, err := dbService.collectionRefStudyParticipant(instanceID, studyKey).UpdateOne(ctx, filter, update)
	return err
}

func (dbService *StudyDBService) CreateMessageScheduledForIndex(instanceID string, studyKey string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_, err := dbService.collectionRefStudyParticipant(instanceID, studyKey).Indexes().CreateOne(
		ctx, mongo.IndexModel{
			Keys: bson.D{
				{Key: "messages.scheduledFor", Value: 1},
				{Key: "studyStatus", Value: 1},
			},
		},
	)
	return err
}

func (dbService *StudyDBService) CreateMessageScheduledForIndexForAllStudies(instanceID string) {
	studies, err := dbService.GetStudiesByStatus(instanceID, "", true)
	if err != nil {
		logger.Error.Printf("unexpected error when fetching studies in '%s': %v", instanceID, err)
		return
	}

	for _, study := range studies {
		err = dbService.CreateMessageScheduledForIndex(instanceID, study.Key)
		if err != nil {
			logger.Error.Printf("unexpected error when creating message schedule indexes: %v", err)
		}
	}
}

func (dbService *StudyDBService) CheckParticipantsForPendingMessages(instanceID string, studyKey string) (hasMessage bool, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{"studyStatus": types.PARTICIPANT_STUDY_STATUS_ACTIVE}
	filter["messages.scheduledFor"] = bson.M{"$lt": time.Now().Unix()}

	elem := &types.ParticipantState{}

	err = dbService.collectionRefStudyParticipant(instanceID, studyKey).FindOne(ctx, filter).Decode(&elem)
	if err == mongo.ErrNoDocuments {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

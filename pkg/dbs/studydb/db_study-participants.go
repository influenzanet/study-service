package studydb

import (
	"github.com/influenzanet/study-service/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// findParticipantsByStudyStatusDB retrieve all participant states from a study by status (e.g. active)
func (dbService *StudyDBService) FindParticipantsByStudyStatus(instanceID string, studyKey string, studyStatus string, useProjection bool) (pStates []types.ParticipantState, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{"studyStatus": studyStatus}

	var opts *options.FindOptions
	if useProjection {
		projection := bson.D{
			primitive.E{Key: "studyStatus", Value: 1},   // {"secretKey", 1},
			primitive.E{Key: "participantID", Value: 1}, // {"secretKey", 1},
		}
		opts = options.Find().SetProjection(projection)
	}

	cur, err := dbService.collectionRefStudyParticipant(instanceID, studyKey).Find(
		ctx,
		filter,
		opts,
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

// FindParticipantState retrieves the participant state for a given participant from a study
func (dbService *StudyDBService) FindAndExecuteOnParticipantsStates(
	instanceID string,
	studyKey string,
	cbk func(dbService *StudyDBService, p types.ParticipantState, instanceID string, studyKey string) error,
) error {

	ctx, cancel := dbService.getContext()
	defer cancel()
	// Get all active participants

	filter := bson.M{"studyStatus": "active"}
	cur, err := dbService.collectionRefStudyParticipant(instanceID, studyKey).Find(ctx, filter)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		// Update state of every participant
		var pState types.ParticipantState
		if err := cur.Decode(&pState); err != nil {
			continue
		}
		// Perform callback:
		if err := cbk(dbService, pState, instanceID, studyKey); err != nil {
			continue
		}
	}
	if err := cur.Err(); err != nil {
		return err
	}
	return nil
}

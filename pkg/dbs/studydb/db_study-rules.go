package studydb

import (
	"github.com/influenzanet/study-service/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (dbService *StudyDBService) SaveStudyRules(instanceID string, rules types.StudyRules) (types.StudyRules, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	//update study object
	study, err := dbService.GetStudyByStudyKey(instanceID, rules.StudyKey)
	if err != nil {
		return rules, err
	}
	study.Rules = rules.Rules
	//TODO: dies hier bereits in endpoint aktualisieren, remove
	_, err = dbService.UpdateStudyInfo(instanceID, study)
	if err != nil {
		return rules, err
	}

	res, err := dbService.collectionRefStudyRules(instanceID).InsertOne(ctx, rules)
	rules.ID = res.InsertedID.(primitive.ObjectID)
	return rules, err
}

func (dbService *StudyDBService) GetCurrentStudyRules(instanceID string, studyKey string) (*types.StudyRules, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	sortByPublishedDesc = bson.D{
		primitive.E{Key: "uploadedAt", Value: -1},
	}

	filter := bson.M{
		"studyKey": studyKey,
	}

	elem := &types.StudyRules{}
	opts := &options.FindOneOptions{
		Sort: sortByPublishedDesc,
	}

	err := dbService.collectionRefStudyRules(instanceID).FindOne(ctx, filter, opts).Decode(&elem)
	return elem, err
}

func (dbService *StudyDBService) GetStudyRulesHistory(instanceID string, studyKey string) (studyRules []*types.StudyRules, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	sortByPublishedDesc = bson.D{
		primitive.E{Key: "uploadedAt", Value: -1},
	}

	filter := bson.M{}
	if len(studyKey) > 0 {
		filter["studyKey"] = studyKey
	}

	opts := &options.FindOptions{
		Sort: sortByPublishedDesc,
	}

	cur, err := dbService.collectionRefStudyRules(instanceID).Find(
		ctx,
		filter,
		opts,
	)

	if err != nil {
		return studyRules, err
	}

	defer cur.Close(ctx)

	studyRules = []*types.StudyRules{}
	for cur.Next(ctx) {
		var result *types.StudyRules
		err := cur.Decode(&result)
		if err != nil {
			return studyRules, err
		}

		studyRules = append(studyRules, result)
	}
	if err := cur.Err(); err != nil {
		return studyRules, err
	}

	return studyRules, nil
}

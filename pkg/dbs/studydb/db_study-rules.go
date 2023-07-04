package studydb

import (
	"github.com/influenzanet/study-service/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (dbService *StudyDBService) AddStudyRules(instanceID string, rules types.StudyRules) (types.StudyRules, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	res, err := dbService.collectionRefStudyRules(instanceID).InsertOne(ctx, rules)
	rules.ID = res.InsertedID.(primitive.ObjectID)
	return rules, err
}

func (dbService *StudyDBService) GetCurrentStudyRules(instanceID string, studyKey string) (*types.StudyRules, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	sortByPublished := bson.D{
		primitive.E{Key: "uploadedAt", Value: -1},
	}

	filter := bson.M{
		"studyKey": studyKey,
	}

	elem := &types.StudyRules{}
	opts := &options.FindOneOptions{
		Sort: sortByPublished,
	}

	err := dbService.collectionRefStudyRules(instanceID).FindOne(ctx, filter, opts).Decode(&elem)
	return elem, err
}

func (dbService *StudyDBService) GetStudyRulesHistory(instanceID string, studyKey string, pageSize int32, page int32, descending bool, since int64, until int64) (studyRulesHistory []*types.StudyRules, totalCount int32, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	sortBy := 1
	if descending {
		sortBy = -1
	}
	sortByPublished := bson.D{
		primitive.E{Key: "uploadedAt", Value: sortBy},
	}

	filter := bson.M{}
	if len(studyKey) > 0 {
		filter["studyKey"] = studyKey
	}

	opts := &options.FindOptions{
		Sort: sortByPublished,
	}
	if pageSize > 0 && page > 0 {
		opts.SetSkip((int64(page) - 1) * int64(pageSize))
		opts.SetLimit(int64(pageSize))
	}
	if since > 0 && until > 0 {
		filter["$and"] = bson.A{
			bson.M{"uploadedAt": bson.M{"$gte": since}},
			bson.M{"uploadedAt": bson.M{"$lte": until}},
		}
	} else if since > 0 {
		filter["uploadedAt"] = bson.M{"$gte": since}
	} else if until > 0 {
		filter["uploadedAt"] = bson.M{"$lte": until}
	}

	cur, err := dbService.collectionRefStudyRules(instanceID).Find(
		ctx,
		filter,
		opts,
	)
	if err != nil {
		return studyRulesHistory, 0, err
	}

	count, err := dbService.collectionRefStudyRules(instanceID).CountDocuments(
		ctx,
		filter,
	)
	totalCount = int32(count)
	if err != nil {
		return studyRulesHistory, 0, err
	}

	defer cur.Close(ctx)

	studyRulesHistory = []*types.StudyRules{}
	for cur.Next(ctx) {
		var result *types.StudyRules
		err := cur.Decode(&result)
		if err != nil {
			return studyRulesHistory, totalCount, err
		}

		studyRulesHistory = append(studyRulesHistory, result)
	}
	if err := cur.Err(); err != nil {
		return studyRulesHistory, totalCount, err
	}

	return studyRulesHistory, totalCount, nil
}

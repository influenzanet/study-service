package studydb

import (
	"github.com/influenzanet/study-service/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (dbService *StudyDBService) SaveStudyRule(instanceID string, rule types.StudyRule) (types.StudyRule, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	//update study object
	study, err := dbService.GetStudyByStudyKey(instanceID, rule.StudyKey)
	if err != nil {
		return rule, err
	}
	study.Rules = rule.Rules
	_, err = dbService.UpdateStudyInfo(instanceID, study)
	if err != nil {
		return rule, err
	}

	res, err := dbService.collectionRefStudyRules(instanceID).InsertOne(ctx, rule)
	rule.ID = res.InsertedID.(primitive.ObjectID)
	return rule, err
}

func (dbService *StudyDBService) GetCurrentStudyRule(instanceID string, studyKey string) (*types.StudyRule, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	sortByPublishedDesc = bson.D{
		primitive.E{Key: "uploadedAt", Value: -1},
	}

	filter := bson.M{
		"studyKey": studyKey,
	}

	elem := &types.StudyRule{}
	opts := &options.FindOneOptions{
		Sort: sortByPublishedDesc,
	}

	err := dbService.collectionRefStudyRules(instanceID).FindOne(ctx, filter, opts).Decode(&elem)
	return elem, err
}

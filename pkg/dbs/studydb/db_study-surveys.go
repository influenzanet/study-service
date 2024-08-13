package studydb

import (
	"errors"
	"time"

	"github.com/coneno/logger"
	"github.com/influenzanet/study-service/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	sortByPublishedDesc = bson.D{
		primitive.E{Key: "published", Value: -1},
	}

	projectionToRemoveSurveyContentAndRules = bson.D{
		primitive.E{Key: "surveyDefinition.items", Value: 0},
		primitive.E{Key: "prefillRules", Value: 0},
		primitive.E{Key: "contextRules", Value: 0},
	}
)

func (dbService *StudyDBService) CreateSurveyDefintionIndexForAllStudies(instanceID string) {
	studies, err := dbService.GetStudiesByStatus(instanceID, "", true)
	if err != nil {
		logger.Error.Printf("unexpected error when fetching studies in '%s': %v", instanceID, err)
		return
	}

	for _, study := range studies {
		err = dbService.CreateSurveyDefintionIndexForStudy(instanceID, study.Key)
		if err != nil {
			logger.Error.Printf("unexpected error when creating survey definition indexes: %v", err)
		}
	}
}

func (dbService *StudyDBService) CreateSurveyDefintionIndexForStudy(instanceID string, studyKey string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_, err := dbService.collectionRefStudySurveys(instanceID, studyKey).Indexes().CreateMany(
		ctx, []mongo.IndexModel{
			{
				Keys: bson.D{
					{Key: "surveyDefinition.key", Value: 1},
					{Key: "unpublished", Value: 1},
					{Key: "published", Value: -1},
				},
			},
			{
				Keys: bson.D{
					{Key: "published", Value: 1},
					{Key: "surveyDefinition.key", Value: 1},
				},
			},
			{
				Keys: bson.D{
					{Key: "unpublished", Value: 1},
				},
			},
			{
				Keys: bson.D{
					{Key: "surveyDefinition.key", Value: 1},
					{Key: "versionID", Value: 1},
				},
				Options: options.Index().SetUnique(true),
			},
		},
	)
	return err
}

func (dbService *StudyDBService) SaveSurvey(instanceID string, studyKey string, survey types.Survey) (types.Survey, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	res, err := dbService.collectionRefStudySurveys(instanceID, studyKey).InsertOne(ctx, survey)
	survey.ID = res.InsertedID.(primitive.ObjectID)
	return survey, err
}

func (dbService *StudyDBService) UnpublishSurvey(instanceID string, studyKey string, surveyKey string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{
		"surveyDefinition.key": surveyKey,
		"unpublished":          0,
	}
	update := bson.M{"$set": bson.M{"unpublished": time.Now().Unix()}}
	_, err := dbService.collectionRefStudySurveys(instanceID, studyKey).UpdateMany(ctx, filter, update)
	return err
}

func (dbService *StudyDBService) GetSurveyKeysInStudy(instanceID string, studyKey string, includeUnpublished bool) (surveyKeys []string, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{}
	if !includeUnpublished {
		filter["unpublished"] = 0
	}
	res, err := dbService.collectionRefStudySurveys(instanceID, studyKey).Distinct(ctx, "surveyDefinition.key", filter)
	if err != nil {
		return surveyKeys, err
	}
	surveyKeys = make([]string, len(res))
	for i, r := range res {
		surveyKeys[i] = r.(string)
	}
	return surveyKeys, err

}

func (dbService *StudyDBService) FindAllCurrentSurveyDefsForStudy(instanceID string, studyKey string, onlyInfos bool) (surveys []*types.Survey, err error) {
	surveyKeys, err := dbService.GetSurveyKeysInStudy(instanceID, studyKey, false)
	if err != nil {
		return nil, err
	}
	for _, key := range surveyKeys {
		survey, err := dbService.FindCurrentSurveyDef(instanceID, studyKey, key, onlyInfos)
		if err != nil {
			logger.Error.Println(err)
			return nil, err
		}
		surveys = append(surveys, survey)
	}
	return surveys, nil
}

func (dbService *StudyDBService) FindCurrentSurveyDef(instanceID string, studyKey string, surveyKey string, onlyInfos bool) (surveys *types.Survey, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{
		"surveyDefinition.key": surveyKey,
		"$or": []bson.M{
			{"unpublished": 0},
			{"unpublished": bson.M{"$exists": false}},
		},
	}

	elem := &types.Survey{}
	opts := &options.FindOneOptions{
		Sort: sortByPublishedDesc,
	}
	if onlyInfos {
		opts.SetProjection(projectionToRemoveSurveyContentAndRules)
	}

	err = dbService.collectionRefStudySurveys(instanceID, studyKey).FindOne(ctx, filter, opts).Decode(&elem)
	return elem, err
}

func (dbService *StudyDBService) FindSurveyDefByVersionID(instanceID string, studyKey string, surveyKey string, versionID string) (surveys *types.Survey, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{
		"surveyDefinition.key": surveyKey,
		"versionID":            versionID,
	}

	elem := &types.Survey{}
	err = dbService.collectionRefStudySurveys(instanceID, studyKey).FindOne(ctx, filter).Decode(&elem)
	return elem, err
}

func (dbService *StudyDBService) DeleteSurveyVersion(instanceID string, studyKey string, surveyKey string, versionID string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{
		"surveyDefinition.key": surveyKey,
		"versionID":            versionID,
	}
	res, err := dbService.collectionRefStudySurveys(instanceID, studyKey).DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if res.DeletedCount < 1 {
		return errors.New("no item was deleted")
	}
	return nil
}

func (dbService *StudyDBService) FindSurveyDefHistory(instanceID string, studyKey string, surveyKey string, onlyInfos bool) (surveys []*types.Survey, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{}
	if len(surveyKey) > 0 {
		filter["surveyDefinition.key"] = surveyKey
	}

	opts := &options.FindOptions{}
	if onlyInfos {
		opts.SetProjection(projectionToRemoveSurveyContentAndRules)
	}

	opts.SetSort(sortByPublishedDesc)

	cur, err := dbService.collectionRefStudySurveys(instanceID, studyKey).Find(
		ctx,
		filter,
		opts,
	)

	if err != nil {
		return surveys, err
	}
	defer cur.Close(ctx)

	surveys = []*types.Survey{}
	for cur.Next(ctx) {
		var result *types.Survey
		err := cur.Decode(&result)
		if err != nil {
			return surveys, err
		}

		surveys = append(surveys, result)
	}
	if err := cur.Err(); err != nil {
		return surveys, err
	}

	return surveys, nil
}

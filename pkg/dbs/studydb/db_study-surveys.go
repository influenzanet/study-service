package studydb

import (
	"errors"
	"time"

	"github.com/coneno/logger"
	"github.com/influenzanet/study-service/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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

func (dbService *StudyDBService) FindAllCurrentSurveyDefsForStudy(instanceID string, studyKey string, includeUnpublished bool, onlyInfos bool) (surveys []*types.Survey, err error) {
	surveyKeys, err := dbService.GetSurveyKeysInStudy(instanceID, studyKey, includeUnpublished)
	if err != nil {
		return surveys, err
	}
	for _, key := range surveyKeys {
		survey, err := dbService.FindCurentSurveyDef(instanceID, studyKey, key, onlyInfos)
		if err != nil {
			logger.Error.Println(err)
			continue
		}
		surveys = append(surveys, survey)
	}
	return surveys, nil
}

func (dbService *StudyDBService) FindCurentSurveyDef(instanceID string, studyKey string, surveyKey string, onlyInfos bool) (surveys *types.Survey, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{
		"surveyDefinition.key": surveyKey,
		"unpublished":          0,
	}

	elem := &types.Survey{}
	opts := &options.FindOneOptions{
		Sort: bson.D{
			primitive.E{Key: "published", Value: -1},
		},
	}
	if onlyInfos {
		projection := bson.D{
			primitive.E{Key: "surveyDefinition.items", Value: 0},
			primitive.E{Key: "prefillRules", Value: 0},
			primitive.E{Key: "contextRules", Value: 0},
		}
		opts.SetProjection(projection)
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
	if res.DeletedCount < 1 {
		err = errors.New("not found")
	}
	return err
}

func (dbService *StudyDBService) FindSurveyDefHistory(instanceID string, studyKey string, surveyKey string, onlyInfos bool) (surveys []*types.Survey, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{}
	if len(surveyKey) > 0 {
		filter["surveyDefinition.key"] = surveyKey
	}

	var opts *options.FindOptions
	if onlyInfos {
		projection := bson.D{
			primitive.E{Key: "surveyDefinition.items", Value: 0},
			primitive.E{Key: "prefillRules", Value: 0},
			primitive.E{Key: "contextRules", Value: 0},
		}
		opts = options.Find().SetProjection(projection)
	}

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

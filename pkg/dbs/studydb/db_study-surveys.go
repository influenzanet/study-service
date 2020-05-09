package studydb

import (
	"errors"

	"github.com/influenzanet/study-service/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (dbService *StudyDBService) SaveSurvey(instanceID string, studyKey string, survey models.Survey) (models.Survey, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{"current.surveyDefinition.key": survey.Current.SurveyDefinition.Key}

	upsert := true
	rd := options.After
	options := options.FindOneAndReplaceOptions{
		Upsert:         &upsert,
		ReturnDocument: &rd,
	}
	elem := models.Survey{}
	err := dbService.collectionRefStudySurveys(instanceID, studyKey).FindOneAndReplace(
		ctx, filter, survey, &options,
	).Decode(&elem)
	return elem, err
}

func (dbService *StudyDBService) RemoveSurveyFromStudy(instanceID string, studyKey string, surveyKey string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{"current.surveyDefinition.key": surveyKey}
	res, err := dbService.collectionRefStudySurveys(instanceID, studyKey).DeleteOne(ctx, filter)
	if res.DeletedCount < 1 {
		err = errors.New("not found")
	}
	return err
}

func (dbService *StudyDBService) FindSurveyDef(instanceID string, studyKey string, surveyKey string) (models.Survey, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{"current.surveyDefinition.key": surveyKey}

	elem := models.Survey{}
	err := dbService.collectionRefStudySurveys(instanceID, studyKey).FindOne(ctx, filter).Decode(&elem)
	return elem, err
}

func (dbService *StudyDBService) FindAllSurveyDefsForStudy(instanceID string, studyKey string, onlyInfos bool) (surveys []models.Survey, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{}

	var opts *options.FindOptions
	if onlyInfos {
		projection := bson.D{
			primitive.E{Key: "key", Value: 1},
			primitive.E{Key: "name", Value: 1},
			primitive.E{Key: "description", Value: 1},
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

	surveys = []models.Survey{}
	for cur.Next(ctx) {
		var result models.Survey
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

package main

import (
	"context"

	"github.com/influenzanet/study-service/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func getStudySecretKey(instanceID string, studyKey string) (secretKey string, err error) {
	projection := bson.D{
		primitive.E{Key: "secretKey", Value: 1}, // {"secretKey", 1},
	}

	var study models.Study
	if err = collectionRefStudyInfos(instanceID).FindOne(
		context.Background(),
		bson.D{
			primitive.E{Key: "studyKey", Value: studyKey}, //{"studyKey", studyKey},
		},
		options.FindOne().SetProjection(projection),
	).Decode(&study); err != nil {
		return "", err
	}
	return study.SecretKey, nil
}

func getStudyMembers(instanceID string, studyKey string) (members []models.StudyMember, err error) {
	projection := bson.D{
		primitive.E{Key: "members", Value: 1}, // {"members", 1},
	}

	var study models.Study
	if err = collectionRefStudyInfos(instanceID).FindOne(
		context.Background(),
		bson.D{
			primitive.E{Key: "studyKey", Value: studyKey}, //{"studyKey", studyKey},
		},
		options.FindOne().SetProjection(projection),
	).Decode(&study); err != nil {
		return []models.StudyMember{}, err
	}
	return study.Members, nil
}

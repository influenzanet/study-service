package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Survey struct {
	ID      primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Current SurveyVersion      `bson:"current"`
	History []SurveyVersion    `bson:"history"`
}

type SurveyVersion struct {
	Published        int64      `bson:"published"`
	UnPublished      int64      `bson:"unpublished"`
	SurveyDefinition SurveyItem `bson:"surveyDefinition"`
}

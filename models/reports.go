package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type SurveyResponseReport struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	SubmittedAt int64              `bson:"submittedAt"`
	By          string             `bson:"by"`
	For         string             `bson:"for"`
	Responses   SurveyItemResponse `bson:"responses"`
}

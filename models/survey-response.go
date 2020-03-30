package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type SurveyResponse struct {
	ID           primitive.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	Key          string               `bson:"key"`
	SubmittedBy  string               `bson:"submittedBy"`
	SubmittedFor string               `bson:"submittedFor"`
	SubmittedAt  int64                `bson:"submittedAt"`
	Responses    []SurveyItemResponse `bson:"responses"`
}

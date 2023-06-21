package types

import "go.mongodb.org/mongo-driver/bson/primitive"

type StudyRules struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	StudyKey      string             `bson:"studyKey"`
	UploadedAt    int64              `bson:"uploadedAt"`
	UploadingUser string             `bson:"uploadingUser"`
	Rules         []Expression       `bson:"rules"`
}

package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Study struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Key       string             `bson:"key"`
	SecretKey string             `bson:"secretKey"`
	Status    string             `bson:"status"`
	Members   []StudyMember      `bson:"members"` // users with access to manage study
	Rules     []Expression       `bson:"rules"`   // defining how the study should run
}

type StudyMember struct {
	UserID string `bson:"key"`
	Role   string `bson:"role"`
}

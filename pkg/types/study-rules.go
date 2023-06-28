package types

import (
	"github.com/influenzanet/study-service/pkg/api"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StudyRules struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	StudyKey   string             `bson:"studyKey"`
	UploadedAt int64              `bson:"uploadedAt"`
	UploadedBy string             `bson:"uploadingUser"`
	Rules      []Expression       `bson:"rules"`
}

func (sr StudyRules) ToAPI() *api.StudyRules {
	rules := make([]*api.Expression, len(sr.Rules))
	for i, r := range sr.Rules {
		rules[i] = r.ToAPI()
	}

	return &api.StudyRules{
		Id:         sr.ID.Hex(),
		StudyKey:   sr.StudyKey,
		UploadedAt: sr.UploadedAt,
		UploadedBy: sr.UploadedBy,
		Rules:      rules,
	}
}

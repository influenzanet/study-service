package models

import (
	"github.com/influenzanet/study-service/api"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SurveyResponse struct {
	ID            primitive.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	Key           string               `bson:"key"`
	ParticipantID string               `bson:"participantID"`
	SubmittedAt   int64                `bson:"submittedAt"`
	Responses     []SurveyItemResponse `bson:"responses"`
	Context       map[string]string    `bson:"context"`
}

func (sr SurveyResponse) ToAPI() *api.SurveyResponse {
	resp := make([]*api.SurveyItemResponse, len(sr.Responses))
	for i, r := range sr.Responses {
		resp[i] = r.ToAPI()
	}
	return &api.SurveyResponse{
		Key:           sr.Key,
		ParticipantId: sr.ParticipantID,
		SubmittedAt:   sr.SubmittedAt,
		Responses:     resp,
		Context:       sr.Context,
	}
}

func SurveyResponseFromAPI(sr *api.SurveyResponse) SurveyResponse {
	if sr == nil {
		return SurveyResponse{}
	}
	resp := make([]SurveyItemResponse, len(sr.Responses))
	for i, r := range sr.Responses {
		resp[i] = SurveyItemResponseFromAPI(r)
	}
	return SurveyResponse{
		Key:           sr.Key,
		ParticipantID: sr.ParticipantId,
		SubmittedAt:   sr.SubmittedAt,
		Responses:     resp,
		Context:       sr.Context,
	}
}

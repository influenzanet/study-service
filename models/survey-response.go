package models

import (
	"github.com/influenzanet/study-service/api"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SurveyResponse struct {
	ID           primitive.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	Key          string               `bson:"key"`
	SubmittedBy  string               `bson:"submittedBy"`
	SubmittedFor string               `bson:"submittedFor"`
	SubmittedAt  int64                `bson:"submittedAt"`
	Responses    []SurveyItemResponse `bson:"responses"`
}

func (sr SurveyResponse) ToAPI() *api.SurveyResponse {
	resp := make([]*api.SurveyItemResponse, len(sr.Responses))
	for i, r := range sr.Responses {
		resp[i] = r.ToAPI()
	}
	return &api.SurveyResponse{
		Key:          sr.Key,
		SubmittedBy:  sr.SubmittedBy,
		SubmittedFor: sr.SubmittedFor,
		SubmittedAt:  sr.SubmittedAt,
		Responses:    resp,
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
		Key:          sr.Key,
		SubmittedBy:  sr.SubmittedBy,
		SubmittedFor: sr.SubmittedFor,
		SubmittedAt:  sr.SubmittedAt,
		Responses:    resp,
	}
}

package types

import (
	"github.com/influenzanet/study-service/pkg/api"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SurveyResponse struct {
	ID            primitive.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	Key           string               `bson:"key" json:"key"`
	ParticipantID string               `bson:"participantID" json:"participantId"`
	VersionID     string               `bson:"versionID" json:"versionId"`
	OpenedAt      int64                `bson:"openedAt" json:"openedAt"`
	SubmittedAt   int64                `bson:"submittedAt" json:"submittedAt"`
	ArrivedAt     int64                `bson:"arrivedAt" json:"arrivedAt"`
	Responses     []SurveyItemResponse `bson:"responses" json:"responses"`
	Context       map[string]string    `bson:"context"  json:"context"`
}

func (sr SurveyResponse) ToAPI() *api.SurveyResponse {
	resp := make([]*api.SurveyItemResponse, len(sr.Responses))
	for i, r := range sr.Responses {
		resp[i] = r.ToAPI()
	}
	return &api.SurveyResponse{
		Id:            sr.ID.Hex(),
		Key:           sr.Key,
		ParticipantId: sr.ParticipantID,
		OpenedAt:      sr.OpenedAt,
		SubmittedAt:   sr.SubmittedAt,
		Responses:     resp,
		VersionId:     sr.VersionID,
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

	_id, _ := primitive.ObjectIDFromHex(sr.Id)
	return SurveyResponse{
		ID:            _id,
		Key:           sr.Key,
		ParticipantID: sr.ParticipantId,
		OpenedAt:      sr.OpenedAt,
		SubmittedAt:   sr.SubmittedAt,
		VersionID:     sr.VersionId,
		Responses:     resp,
		Context:       sr.Context,
	}
}

package types

import (
	api "github.com/influenzanet/study-service/pkg/api"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Report struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Key           string             `bson:"key" json:"key"`
	ParticipantID string             `bson:"participantID" json:"participantID"` // reference to the study specific participant ID
	ResponseID    string             `bson:"responseID" json:"responseID"`       // reference to the report
	Timestamp     int64              `bson:"timestamp" json:"timestamp"`
	Data          []ReportData       `bson:"data" json:"data"`
}

type ReportData struct {
	Key   string `bson:"key" json:"key"`
	Value string `bson:"value" json:"value"`
	Dtype string `bson:"dtype,omitempty" json:"dtype,omitempty"`
}

func (r Report) ToAPI() *api.Report {
	data := make([]*api.Report_Data, len(r.Data))
	for i, ea := range r.Data {
		data[i] = ea.ToAPI()
	}
	return &api.Report{
		Id:            r.ID.Hex(),
		Key:           r.Key,
		ParticipantId: r.ParticipantID,
		ResponseId:    r.ResponseID,
		Timestamp:     r.Timestamp,
		Data:          data,
	}
}

func (r ReportData) ToAPI() *api.Report_Data {
	return &api.Report_Data{
		Key:   r.Key,
		Value: r.Value,
		Dtype: r.Dtype,
	}
}

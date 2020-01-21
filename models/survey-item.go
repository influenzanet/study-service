package models

import (
	api "github.com/influenzanet/study-service/api"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SurveyItem struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Key         string             `bson:"key"`
	Follows     []string           `bson:"follows"`
	Condition   Expression         `bson:"condition"`
	Priority    float32            `bson:"priority"`
	Version     int32              `bson:"version"`
	VersionTags []string           `bson:"versionTags"`

	// Question group attributes ->
	Items           []SurveyItem `bson:"items"`
	SelectionMethod Expression   `bson:"selectionMethod"`

	// Question attributes ->
	Type string `bson:"type"`
	// repeated Component components = 10;
	// repeated Validation validations = 11;
}

func (s SurveyItem) ToAPI() *api.SurveyItem {
	items := make([]*api.SurveyItem, len(s.Items))
	for i, si := range s.Items {
		items[i] = si.ToAPI()
	}

	return &api.SurveyItem{
		Id:              s.ID.Hex(),
		Key:             s.Key,
		Follows:         s.Follows,
		Condition:       s.Condition.ToAPI(),
		Priority:        s.Priority,
		Version:         s.Version,
		VersionTags:     s.VersionTags,
		Items:           items,
		SelectionMethod: s.SelectionMethod.ToAPI(),
		Type:            s.Type,
	}
}

func SurveyItemFromAPI(s *api.SurveyItem) SurveyItem {
	if s == nil {
		return SurveyItem{}
	}
	items := make([]SurveyItem, len(s.Items))
	for i, si := range s.Items {
		items[i] = SurveyItemFromAPI(si)
	}

	_id, _ := primitive.ObjectIDFromHex(s.Id)

	return SurveyItem{
		ID:              _id,
		Key:             s.Key,
		Follows:         s.Follows,
		Condition:       ExpressionFromAPI(s.Condition),
		Priority:        s.Priority,
		Version:         s.Version,
		VersionTags:     s.VersionTags,
		Items:           items,
		SelectionMethod: ExpressionFromAPI(s.SelectionMethod),
		Type:            s.Type,
	}
}

package models

import (
	api "github.com/influenzanet/study-service/api"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SurveyItem struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Key         string             `bson:"key"`
	Follows     []string           `bson:"follows,omitempty"`
	Condition   *Expression        `bson:"condition,omitempty"`
	Priority    float32            `bson:"priority,omitempty"`
	Version     int32              `bson:"version,omitempty"`
	VersionTags []string           `bson:"versionTags,omitempty"`

	// Question group attributes ->
	Items           []SurveyItem `bson:"items,omitempty"`
	SelectionMethod *Expression  `bson:"selectionMethod,omitempty"`

	// Question attributes ->
	Type        string        `bson:"type,omitempty"`
	Components  ItemComponent `bson:"components,omitempty"`
	Validations []Validation  `bson:"validations,omitempty"`
}

func (s SurveyItem) ToAPI() *api.SurveyItem {
	items := make([]*api.SurveyItem, len(s.Items))
	for i, si := range s.Items {
		items[i] = si.ToAPI()
	}

	validations := make([]*api.Validation, len(s.Validations))
	for i, si := range s.Validations {
		validations[i] = si.ToAPI()
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
		Components:      s.Components.ToAPI(),
		Validations:     validations,
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

	validations := make([]Validation, len(s.Validations))
	for i, si := range s.Validations {
		validations[i] = ValidationFromAPI(si)
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
		Components:      ItemComponentFromAPI(s.Components),
		Validations:     validations,
	}
}

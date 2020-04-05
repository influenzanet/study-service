package models

import (
	api "github.com/influenzanet/study-service/api"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Survey struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name        []LocalisedObject  `bson:"name"`
	Description []LocalisedObject  `bson:"description"`
	Current     SurveyVersion      `bson:"current"`
	History     []SurveyVersion    `bson:"history"`
}

type SurveyVersion struct {
	Published        int64      `bson:"published"`
	UnPublished      int64      `bson:"unpublished"`
	SurveyDefinition SurveyItem `bson:"surveyDefinition"`
}

func (s Survey) ToAPI() *api.Survey {
	history := make([]*api.SurveyVersion, len(s.History))
	for i, si := range s.History {
		history[i] = si.ToAPI()
	}
	name := make([]*api.LocalisedObject, len(s.Name))
	for i, si := range s.Name {
		name[i] = si.ToAPI()
	}
	description := make([]*api.LocalisedObject, len(s.Description))
	for i, si := range s.Description {
		description[i] = si.ToAPI()
	}
	return &api.Survey{
		Id:          s.ID.Hex(),
		Name:        name,
		Description: description,
		Current:     s.Current.ToAPI(),
		History:     history,
	}
}

func SurveyFromAPI(s *api.Survey) Survey {
	if s == nil {
		return Survey{}
	}
	_id, _ := primitive.ObjectIDFromHex(s.Id)

	history := make([]SurveyVersion, len(s.History))
	for i, si := range s.History {
		history[i] = SurveyVersionFromAPI(si)
	}
	name := make([]LocalisedObject, len(s.Name))
	for i, si := range s.Name {
		name[i] = LocalisedObjectFromAPI(si)
	}
	description := make([]LocalisedObject, len(s.Description))
	for i, si := range s.Description {
		description[i] = LocalisedObjectFromAPI(si)
	}
	return Survey{
		ID:          _id,
		Name:        name,
		Description: description,
		Current:     SurveyVersionFromAPI(s.Current),
		History:     history,
	}
}

func (sv SurveyVersion) ToAPI() *api.SurveyVersion {
	return &api.SurveyVersion{
		Published:        sv.Published,
		Unpublished:      sv.UnPublished,
		SurveyDefinition: sv.SurveyDefinition.ToAPI(),
	}
}

func SurveyVersionFromAPI(sv *api.SurveyVersion) SurveyVersion {
	if sv == nil {
		return SurveyVersion{}
	}
	return SurveyVersion{
		Published:        sv.Published,
		UnPublished:      sv.Unpublished,
		SurveyDefinition: SurveyItemFromAPI(sv.SurveyDefinition),
	}
}

package types

import (
	api "github.com/influenzanet/study-service/pkg/api"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	SURVEY_AVAILABLE_FOR_PUBLIC                 = "public"
	SURVEY_AVAILABLE_FOR_TEMPORARY_PARTICIPANTS = "temporary_participants"
	SURVEY_AVAILABLE_FOR_ACTIVE_PARTICIPANTS    = "active_participants"
)

type SurveyVersionsJSON struct {
	SurveyVersions []*Survey `json:"surveyVersions"`
}

type Survey struct {
	ID                           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Props                        SurveyProps        `bson:"props,omitempty"`
	PrefillRules                 []Expression       `bson:"prefillRules,omitempty"`
	ContextRules                 *SurveyContextDef  `bson:"contextRules,omitempty"`
	MaxItemsPerPage              *MaxItemsPerPage   `bson:"maxItemsPerPage,omitempty"`
	AvailableFor                 string             `bson:"availableFor,omitempty"`
	RequireLoginBeforeSubmission bool               `bson:"requireLoginBeforeSubmission"`

	Published        int64             `bson:"published"`
	Unpublished      int64             `bson:"unpublished"`
	SurveyDefinition SurveyItem        `bson:"surveyDefinition"`
	VersionID        string            `bson:"versionID"`
	Metadata         map[string]string `bson:"metadata,omitempty"`
}

type SurveyProps struct {
	Name            []LocalisedObject `bson:"name"`
	Description     []LocalisedObject `bson:"description"`
	TypicalDuration []LocalisedObject `bson:"typicalDuration"`
}

type MaxItemsPerPage struct {
	Large int32 `bson:"large"`
	Small int32 `bson:"small"`
}

func (s Survey) ToAPI() *api.Survey {
	prefills := make([]*api.Expression, len(s.PrefillRules))
	for i, r := range s.PrefillRules {
		prefills[i] = r.ToAPI()
	}
	as := &api.Survey{
		Id:               s.ID.Hex(),
		Props:            s.Props.ToAPI(),
		PrefillRules:     prefills,
		Published:        s.Published,
		Unpublished:      s.Unpublished,
		SurveyDefinition: s.SurveyDefinition.ToAPI(),
		VersionId:        s.VersionID,
		Metadata:         s.Metadata,
	}
	if s.ContextRules != nil {
		as.ContextRules = s.ContextRules.ToAPI()
	}
	if s.MaxItemsPerPage != nil {
		as.MaxItemsPerPage = s.MaxItemsPerPage.ToAPI()
	}
	as.AvailableFor = s.AvailableFor
	as.RequireLoginBeforeSubmission = s.RequireLoginBeforeSubmission
	return as
}

func SurveyFromAPI(s *api.Survey) Survey {
	if s == nil {
		return Survey{}
	}
	_id, _ := primitive.ObjectIDFromHex(s.Id)

	prefills := make([]Expression, len(s.PrefillRules))
	for i, r := range s.PrefillRules {
		prefills[i] = *ExpressionFromAPI(r)
	}
	return Survey{
		ID:                           _id,
		Props:                        Survey_PropsFromAPI(s.Props),
		PrefillRules:                 prefills,
		ContextRules:                 SurveyContextDefFromAPI(s.ContextRules),
		MaxItemsPerPage:              MaxItemsPerPageFromAPI(s.MaxItemsPerPage),
		AvailableFor:                 s.AvailableFor,
		RequireLoginBeforeSubmission: s.RequireLoginBeforeSubmission,
		Published:                    s.Published,
		Unpublished:                  s.Unpublished,
		SurveyDefinition:             SurveyItemFromAPI(s.SurveyDefinition),
		VersionID:                    s.VersionId,
		Metadata:                     s.Metadata,
	}
}

func (s MaxItemsPerPage) ToAPI() *api.MaxItemsPerPage {
	return &api.MaxItemsPerPage{
		Large: s.Large,
		Small: s.Small,
	}
}

func MaxItemsPerPageFromAPI(s *api.MaxItemsPerPage) *MaxItemsPerPage {
	if s == nil {
		return nil
	}
	return &MaxItemsPerPage{
		Large: s.Large,
		Small: s.Small,
	}
}

func (sp SurveyProps) ToAPI() *api.Survey_Props {
	name := make([]*api.LocalisedObject, len(sp.Name))
	for i, r := range sp.Name {
		name[i] = r.ToAPI()
	}
	description := make([]*api.LocalisedObject, len(sp.Description))
	for i, r := range sp.Description {
		description[i] = r.ToAPI()
	}

	td := make([]*api.LocalisedObject, len(sp.TypicalDuration))
	for i, r := range sp.TypicalDuration {
		td[i] = r.ToAPI()
	}
	return &api.Survey_Props{
		Name:            name,
		Description:     description,
		TypicalDuration: td,
	}
}

func Survey_PropsFromAPI(sp *api.Survey_Props) SurveyProps {
	if sp == nil {
		return SurveyProps{}
	}
	name := make([]LocalisedObject, len(sp.Name))
	for i, r := range sp.Name {
		name[i] = LocalisedObjectFromAPI(r)
	}
	description := make([]LocalisedObject, len(sp.Description))
	for i, r := range sp.Description {
		description[i] = LocalisedObjectFromAPI(r)
	}
	td := make([]LocalisedObject, len(sp.TypicalDuration))
	for i, r := range sp.TypicalDuration {
		td[i] = LocalisedObjectFromAPI(r)
	}
	return SurveyProps{
		Name:            name,
		Description:     description,
		TypicalDuration: td,
	}
}

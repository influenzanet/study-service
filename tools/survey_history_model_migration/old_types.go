package main

import (
	"fmt"
	"time"

	"github.com/coneno/logger"
	"github.com/influenzanet/study-service/pkg/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OldSurvey struct {
	ID                           primitive.ObjectID      `bson:"_id,omitempty" json:"id,omitempty"`
	Props                        types.SurveyProps       `bson:"props,omitempty"`
	Current                      OldSurveyVersion        `bson:"current"`
	History                      []OldSurveyVersion      `bson:"history,omitempty"`
	PrefillRules                 []types.Expression      `bson:"prefillRules,omitempty"`
	ContextRules                 *types.SurveyContextDef `bson:"contextRules,omitempty"`
	MaxItemsPerPage              *types.MaxItemsPerPage  `bson:"maxItemsPerPage,omitempty"`
	AvailableFor                 string                  `bson:"availableFor,omitempty"`
	RequireLoginBeforeSubmission bool                    `bson:"requireLoginBeforeSubmission"`
}

func (s OldSurvey) ToNew() []*types.Survey {
	surveyVersions := []*types.Survey{}

	// add history:
	for i, sv := range s.History {
		version := sv.VersionID
		if len(version) < 1 {
			version = fmt.Sprintf("v%d", i)
			logger.Info.Printf("'%s' history object unpublished at %s had no versionId, use %s instead", s.Current.SurveyDefinition.Key, time.Unix(sv.UnPublished, 0), version)
		}
		surveyVersions = append(surveyVersions,
			&types.Survey{
				ID:                           s.ID,
				Props:                        s.Props,
				PrefillRules:                 s.PrefillRules,
				ContextRules:                 s.ContextRules,
				MaxItemsPerPage:              s.MaxItemsPerPage,
				AvailableFor:                 s.AvailableFor,
				RequireLoginBeforeSubmission: s.RequireLoginBeforeSubmission,

				Published:        sv.Published,
				Unpublished:      sv.UnPublished,
				SurveyDefinition: sv.SurveyDefinition.ToNew(),
				VersionID:        version,
			},
		)
	}

	// add current:
	version := s.Current.VersionID
	if len(version) < 1 {
		version = fmt.Sprintf("v%d", len(s.History))
		logger.Info.Printf("'%s' current had no versionId, use %s instead", s.Current.SurveyDefinition.Key, version)
	}
	surveyVersions = append(surveyVersions,
		&types.Survey{
			ID:                           s.ID,
			Props:                        s.Props,
			PrefillRules:                 s.PrefillRules,
			ContextRules:                 s.ContextRules,
			MaxItemsPerPage:              s.MaxItemsPerPage,
			AvailableFor:                 s.AvailableFor,
			RequireLoginBeforeSubmission: s.RequireLoginBeforeSubmission,

			Published:        s.Current.Published,
			Unpublished:      s.Current.UnPublished,
			SurveyDefinition: s.Current.SurveyDefinition.ToNew(),
			VersionID:        version,
		},
	)

	return surveyVersions

}

type OldSurveyVersion struct {
	Published        int64         `bson:"published"`
	UnPublished      int64         `bson:"unpublished"`
	SurveyDefinition OldSurveyItem `bson:"surveyDefinition"`
	VersionID        string        `bson:"versionID"`
}

type OldSurveyItem struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Key       string             `bson:"key"`
	Follows   []string           `bson:"follows,omitempty"`
	Condition *types.Expression  `bson:"condition,omitempty"`
	Priority  float32            `bson:"priority,omitempty"`

	// --> this are removed
	Version     int32    `bson:"version,omitempty"`
	VersionTags []string `bson:"versionTags,omitempty"`
	// <--

	// Question group attributes ->
	Items           []OldSurveyItem   `bson:"items,omitempty"`
	SelectionMethod *types.Expression `bson:"selectionMethod,omitempty"`

	// Question attributes ->
	Type             string               `bson:"type,omitempty"` // Specify some special types e.g. 'pageBreak','surveyEnd'
	Components       *types.ItemComponent `bson:"components,omitempty"`
	Validations      []types.Validation   `bson:"validations,omitempty"`
	ConfidentialMode string               `bson:"confidentialMode,omitempty"`
}

func (si OldSurveyItem) ToNew() types.SurveyItem {
	items := make([]types.SurveyItem, len(si.Items))
	for i, item := range si.Items {
		items[i] = item.ToNew()
	}
	return types.SurveyItem{
		ID:        si.ID,
		Key:       si.Key,
		Follows:   si.Follows,
		Condition: si.Condition,
		Priority:  si.Priority,

		Metadata: map[string]string{
			"version": fmt.Sprintf("%d", si.Version),
		},

		Items:           items,
		SelectionMethod: si.SelectionMethod,

		Type:             si.Type,
		Components:       si.Components,
		Validations:      si.Validations,
		ConfidentialMode: si.ConfidentialMode,
	}
}

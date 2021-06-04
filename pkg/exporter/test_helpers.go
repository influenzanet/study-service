package exporter

import (
	"strconv"

	studyAPI "github.com/influenzanet/study-service/pkg/api"
)

func mockQuestion(
	key string,
	lang string,
	title string,
	responseOptions *studyAPI.ItemComponent,
) *studyAPI.SurveyItem {
	q := studyAPI.SurveyItem{
		Key: key,
		Components: &studyAPI.ItemComponent{
			Role: "root",
			Items: []*studyAPI.ItemComponent{
				{Role: "title", Content: []*studyAPI.LocalisedObject{
					{Code: lang, Parts: []*studyAPI.ExpressionArg{{Data: &studyAPI.ExpressionArg_Str{Str: title}}}},
				}},
				responseOptions,
			},
		},
	}
	return &q
}

type MockOpionDef struct {
	Key   string
	Role  string
	Label string
}

func mockSingleChoiceGroup(lang string, options []MockOpionDef) *studyAPI.ItemComponent {
	rg := studyAPI.ItemComponent{
		Key:  "rg",
		Role: "responseGroup", Items: []*studyAPI.ItemComponent{
			{Key: "scg", Role: "singleChoiceGroup", Items: []*studyAPI.ItemComponent{}},
		}}
	for _, o := range options {
		rg.Items[0].Items = append(rg.Items[0].Items,
			&studyAPI.ItemComponent{Key: o.Key, Role: o.Role, Content: []*studyAPI.LocalisedObject{
				{Code: lang, Parts: []*studyAPI.ExpressionArg{{Data: &studyAPI.ExpressionArg_Str{Str: o.Label}}}},
			}},
		)
	}
	return &rg
}

func mockMultipleChoiceGroup(lang string, options []MockOpionDef) *studyAPI.ItemComponent {
	rg := studyAPI.ItemComponent{
		Key:  "rg",
		Role: "responseGroup", Items: []*studyAPI.ItemComponent{
			{Key: "mcg", Role: "multipleChoiceGroup", Items: []*studyAPI.ItemComponent{}},
		}}
	for _, o := range options {
		rg.Items[0].Items = append(rg.Items[0].Items,
			&studyAPI.ItemComponent{Key: o.Key, Role: o.Role, Content: []*studyAPI.LocalisedObject{
				{Code: lang, Parts: []*studyAPI.ExpressionArg{{Data: &studyAPI.ExpressionArg_Str{Str: o.Label}}}},
			}},
		)
	}
	return &rg
}

func mockLikertGroup(lang string, categoryLabels []MockOpionDef, optionLabels []string) *studyAPI.ItemComponent {
	rg := studyAPI.ItemComponent{
		Key:  "rg",
		Role: "responseGroup", Items: []*studyAPI.ItemComponent{
			{Key: "lg", Role: "likertGroup", Items: []*studyAPI.ItemComponent{}},
		}}

	for i, o := range categoryLabels {
		rg.Items[0].Items = append(rg.Items[0].Items,
			&studyAPI.ItemComponent{Key: strconv.Itoa(i), Role: "text", Content: []*studyAPI.LocalisedObject{
				{Code: lang, Parts: []*studyAPI.ExpressionArg{{Data: &studyAPI.ExpressionArg_Str{Str: o.Label}}}},
			}},
		)
		rg.Items[0].Items = append(rg.Items[0].Items,
			&studyAPI.ItemComponent{Key: o.Key, Role: "likert", Items: []*studyAPI.ItemComponent{}},
		)

		index := len(rg.Items[0].Items) - 1
		for j, label := range optionLabels {
			rg.Items[0].Items[index].Items = append(rg.Items[0].Items[index].Items, &studyAPI.ItemComponent{Key: strconv.Itoa(j + 1), Role: "option", Content: []*studyAPI.LocalisedObject{
				{Code: lang, Parts: []*studyAPI.ExpressionArg{{Data: &studyAPI.ExpressionArg_Str{Str: label}}}},
			}})
		}
	}
	return &rg
}

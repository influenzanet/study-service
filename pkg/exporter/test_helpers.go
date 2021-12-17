package exporter

import (
	"strconv"

	"github.com/influenzanet/study-service/pkg/types"
)

func mockQuestion(
	key string,
	lang string,
	title string,
	responseOptions *types.ItemComponent,
) *types.SurveyItem {
	q := types.SurveyItem{
		Key: key,
		Components: &types.ItemComponent{
			Role: "root",
			Items: []types.ItemComponent{
				{Role: "title", Content: []types.LocalisedObject{
					{Code: lang, Parts: []types.ExpressionArg{{Str: title}}},
				}},
				*responseOptions,
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

func mockSingleChoiceGroup(lang string, options []MockOpionDef) *types.ItemComponent {
	rg := types.ItemComponent{
		Key:  "rg",
		Role: "responseGroup", Items: []types.ItemComponent{
			{Key: "scg", Role: "singleChoiceGroup", Items: []types.ItemComponent{}},
		}}
	for _, o := range options {
		rg.Items[0].Items = append(rg.Items[0].Items,
			types.ItemComponent{Key: o.Key, Role: o.Role, Content: []types.LocalisedObject{
				{Code: lang, Parts: []types.ExpressionArg{{Str: o.Label}}},
			}},
		)
	}
	return &rg
}

func mockMultipleChoiceGroup(lang string, options []MockOpionDef) *types.ItemComponent {
	rg := types.ItemComponent{
		Key:  "rg",
		Role: "responseGroup", Items: []types.ItemComponent{
			{Key: "mcg", Role: "multipleChoiceGroup", Items: []types.ItemComponent{}},
		}}
	for _, o := range options {
		rg.Items[0].Items = append(rg.Items[0].Items,
			types.ItemComponent{Key: o.Key, Role: o.Role, Content: []types.LocalisedObject{
				{Code: lang, Parts: []types.ExpressionArg{{Str: o.Label}}},
			}},
		)
	}
	return &rg
}

func mockLikertGroup(lang string, categoryLabels []MockOpionDef, optionLabels []string) *types.ItemComponent {
	rg := types.ItemComponent{
		Key:  "rg",
		Role: "responseGroup", Items: []types.ItemComponent{
			{Key: "lg", Role: "likertGroup", Items: []types.ItemComponent{}},
		}}

	for i, o := range categoryLabels {
		rg.Items[0].Items = append(rg.Items[0].Items,
			types.ItemComponent{Key: strconv.Itoa(i), Role: "text", Content: []types.LocalisedObject{
				{Code: lang, Parts: []types.ExpressionArg{{Str: o.Label}}},
			}},
		)
		rg.Items[0].Items = append(rg.Items[0].Items,
			types.ItemComponent{Key: o.Key, Role: "likert", Items: []types.ItemComponent{}},
		)

		index := len(rg.Items[0].Items) - 1
		for j, label := range optionLabels {
			rg.Items[0].Items[index].Items = append(rg.Items[0].Items[index].Items, types.ItemComponent{Key: strconv.Itoa(j + 1), Role: "option", Content: []types.LocalisedObject{
				{Code: lang, Parts: []types.ExpressionArg{{Str: label}}},
			}})
		}
	}
	return &rg
}

func mockResponsiveSingleChoiceArray(lang string, categoryLabels []MockOpionDef, optionLabels []string) types.ItemComponent {
	rg := types.ItemComponent{
		Key:  "rg",
		Role: "responseGroup", Items: []types.ItemComponent{
			{Key: "rsca", Role: "responsiveSingleChoiceArray", Items: []types.ItemComponent{}},
		}}

	rg.Items[0].Items = append(rg.Items[0].Items, types.ItemComponent{
		Key:  "options",
		Role: "options",
	})
	for j, label := range optionLabels {
		rg.Items[0].Items[0].Items = append(rg.Items[0].Items[0].Items, types.ItemComponent{Key: strconv.Itoa(j + 1), Role: "option", Content: []types.LocalisedObject{
			{Code: lang, Parts: []types.ExpressionArg{{Str: label}}},
		}})
	}

	for _, o := range categoryLabels {
		rg.Items[0].Items = append(rg.Items[0].Items,
			types.ItemComponent{Key: o.Key, Role: "row", Content: []types.LocalisedObject{
				{Code: lang, Parts: []types.ExpressionArg{{Str: o.Label}}},
			}},
		)
	}
	return rg
}

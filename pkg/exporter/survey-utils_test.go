package exporter

import (
	"testing"

	"github.com/influenzanet/study-service/pkg/types"
)

func TestIsItemGroup(t *testing.T) {
	testLang := "en"

	testItem1 := &types.SurveyItem{Key: "weeky.G1", Items: []types.SurveyItem{
		*mockQuestion("weekly.G1.Q1", testLang, "Title of Group 1's Q1", mockLikertGroup(testLang, []MockOpionDef{
			{Key: "cat1", Label: "Category 1"},
			{Key: "cat2", Label: "Category 2"},
		}, []string{
			"o1", "o2", "o3",
		})),
	}}
	testItem2 := mockQuestion("weekly.Q2", testLang, "Title of Q2", mockMultipleChoiceGroup(testLang, []MockOpionDef{
		{Key: "1", Role: "option", Label: "Option 1"},
		{Key: "2", Role: "option", Label: "Option 2"},
		{Key: "3", Role: "input", Label: "Other"},
	}))

	t.Run("with with missing item", func(t *testing.T) {
		if isItemGroup(nil) {
			t.Error("missing item wrongly as group")
		}
	})

	t.Run("with with single item", func(t *testing.T) {
		if isItemGroup(testItem2) {
			t.Error("single item wrongly as group")
		}
	})

	t.Run("with with group item", func(t *testing.T) {
		if !isItemGroup(testItem1) {
			t.Error("group item not recognized")
		}
	})
}

func TestGetResponseGroupComponent(t *testing.T) {
	testLang := "en"

	testItem1 := mockQuestion("weekly.Q2", testLang, "Title of Q2", mockMultipleChoiceGroup(testLang, []MockOpionDef{
		{Key: "1", Role: "option", Label: "Option 1"},
		{Key: "2", Role: "option", Label: "Option 2"},
		{Key: "3", Role: "input", Label: "Other"},
	}))

	t.Run("with test items", func(t *testing.T) {
		rg := getResponseGroupComponent(testItem1)
		if rg == nil {
			t.Error("rg empty")
			return
		}
		if rg.Role != "responseGroup" {
			t.Errorf("unexpected role: %s", rg.Role)
			return
		}
	})
}

func TestGetTranslation(t *testing.T) {

	t.Run("with empty translation list", func(t *testing.T) {
		_, err := getTranslation(&[]types.LocalisedObject{}, "en")
		if err == nil {
			t.Error("should return an error")
			return
		}
		if err.Error() != "translations missing" {
			t.Errorf("unexpected error: %v", err)
			return
		}
	})

	t.Run("with missing translation", func(t *testing.T) {
		_, err := getTranslation(&[]types.LocalisedObject{
			{Code: "de", Parts: []types.ExpressionArg{{DType: "str", Str: "Test DE"}}},
			{Code: "nl", Parts: []types.ExpressionArg{{DType: "str", Str: "Test NL"}}},
		}, "en")
		if err == nil {
			t.Error("should return an error")
			return
		}
		if err.Error() != "translation missing" {
			t.Errorf("unexpected error: %v", err)
			return
		}
	})

	t.Run("with single part", func(t *testing.T) {
		tr, err := getTranslation(&[]types.LocalisedObject{
			{Code: "de", Parts: []types.ExpressionArg{{DType: "str", Str: "Test DE"}}},
			{Code: "en", Parts: []types.ExpressionArg{{DType: "str", Str: "Test EN"}}},
			{Code: "nl", Parts: []types.ExpressionArg{{DType: "str", Str: "Test NL"}}},
		}, "en")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if tr != "Test EN" {
			t.Errorf("unexpected value: %s", tr)
			return
		}
	})

	t.Run("with multiple parts", func(t *testing.T) {
		tr, err := getTranslation(&[]types.LocalisedObject{
			{Code: "de", Parts: []types.ExpressionArg{{DType: "str", Str: "Test DE"}}},
			{Code: "en", Parts: []types.ExpressionArg{
				{DType: "str", Str: "Test "},
				{DType: "exp", Exp: &types.Expression{}},
				{DType: "str", Str: "EN"},
			}},
			{Code: "nl", Parts: []types.ExpressionArg{{DType: "str", Str: "Test NL"}}},
		}, "en")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if tr != "Test EN" {
			t.Errorf("unexpected value: %s", tr)
			return
		}
	})
}

func TestExtractResponses(t *testing.T) {
	testLang := "en"
	t.Run("missing response group component", func(t *testing.T) {
		ro, qType := extractResponses(nil, testLang)
		if len(ro) > 0 {
			t.Error("should be empty")
		}
		if qType != QUESTION_TYPE_EMPTY {
			t.Errorf("unexpected question type: %s", qType)
		}
	})

	t.Run("missing items", func(t *testing.T) {
		rg := types.ItemComponent{
			Key:   "rg",
			Role:  "responseGroup",
			Items: []types.ItemComponent{},
		}
		ro, qType := extractResponses(&rg, testLang)
		if len(ro) > 0 {
			t.Error("should be empty")
		}
		if qType != QUESTION_TYPE_EMPTY {
			t.Errorf("unexpected question type: %s", qType)
		}
	})

	t.Run("multiple items (not known roles)", func(t *testing.T) {
		rg := types.ItemComponent{
			Key:  "rg",
			Role: "responseGroup",
			Items: []types.ItemComponent{
				{Key: "1", Role: "text"},
				{Key: "2", Role: "something"},
				{Key: "3", Role: "more"},
			},
		}
		ro, qType := extractResponses(&rg, testLang)
		if len(ro) > 0 {
			t.Error("should be empty")
		}
		if qType != QUESTION_TYPE_EMPTY {
			t.Errorf("unexpected question type: %s", qType)
		}
	})

	t.Run("single choice group", func(t *testing.T) {
		rg := types.ItemComponent{
			Key:  "rg",
			Role: "responseGroup",
			Items: []types.ItemComponent{
				{Key: "scg", Role: "singleChoiceGroup", Items: []types.ItemComponent{
					{Key: "1", Role: "option"},
					{Key: "2", Role: "text"},
					{Key: "3", Role: "input"},
					{Key: "4", Role: "dateInput"},
				}},
			},
		}
		ro, qType := extractResponses(&rg, testLang)
		if len(ro) != 1 {
			t.Error("shouldn't be empty")
		}
		if qType != QUESTION_TYPE_SINGLE_CHOICE {
			t.Errorf("unexpected question type: %s", qType)
		}
	})

	t.Run("multiple choice group", func(t *testing.T) {
		rg := types.ItemComponent{
			Key:  "rg",
			Role: "responseGroup",
			Items: []types.ItemComponent{
				{Key: "mcg", Role: "multipleChoiceGroup", Items: []types.ItemComponent{
					{Key: "1", Role: "option"},
					{Key: "2", Role: "text"},
					{Key: "3", Role: "input"},
					{Key: "4", Role: "dateInput"},
				}},
			},
		}
		ro, qType := extractResponses(&rg, testLang)
		if len(ro) != 1 {
			t.Error("shouldn't be empty")
		}
		if qType != QUESTION_TYPE_MULTIPLE_CHOICE {
			t.Errorf("unexpected question type: %s", qType)
		}
	})

	t.Run("likert group", func(t *testing.T) {
		rg := types.ItemComponent{
			Key:  "rg",
			Role: "responseGroup",
			Items: []types.ItemComponent{
				{Key: "lg", Role: "likertGroup", Items: []types.ItemComponent{
					{Key: "1", Role: "text"},
					{Key: "2", Role: "likert", Items: []types.ItemComponent{
						{Key: "1", Role: "option"},
						{Key: "2", Role: "option"},
						{Key: "3", Role: "option"},
					}},
					{Key: "3", Role: "text"},
					{Key: "4", Role: "likert", Items: []types.ItemComponent{
						{Key: "1", Role: "option"},
						{Key: "2", Role: "option"},
						{Key: "3", Role: "option"},
					}},
				}},
			},
		}
		ro, qType := extractResponses(&rg, testLang)
		if len(ro) != 2 {
			t.Error("shouldn't be empty")
		}
		if qType != QUESTION_TYPE_LIKERT_GROUP {
			t.Errorf("unexpected question type: %s", qType)
		}
	})

	t.Run("responsive single choice array", func(t *testing.T) {
		rg := mockResponsiveSingleChoiceArray("en", []MockOpionDef{
			{Key: "cat1", Label: "Category 1"},
			{Key: "cat2", Label: "Category 2"},
		}, []string{
			"o1", "o2", "o3",
		})
		ro, qType := extractResponses(&rg, testLang)
		if len(ro) != 2 {
			t.Error("shouldn't be empty")
		}
		if qType != QUESTION_TYPE_RESPONSIVE_SINGLE_CHOICE_ARRAY {
			t.Errorf("unexpected question type: %s", qType)
		}
	})

	t.Run("likerts - but not likertGroup", func(t *testing.T) {
		rg := types.ItemComponent{
			Key:  "rg",
			Role: "responseGroup",
			Items: []types.ItemComponent{

				{Key: "1", Role: "text"},
				{Key: "2", Role: "likert", Items: []types.ItemComponent{
					{Key: "1", Role: "option"},
					{Key: "2", Role: "option"},
					{Key: "3", Role: "option"},
				}},
				{Key: "3", Role: "text"},
				{Key: "4", Role: "likert", Items: []types.ItemComponent{
					{Key: "1", Role: "option"},
					{Key: "2", Role: "option"},
					{Key: "3", Role: "option"},
				}},
			},
		}
		ro, qType := extractResponses(&rg, testLang)
		if len(ro) != 2 {
			t.Error("shouldn't be empty")
		}
		if qType != QUESTION_TYPE_LIKERT {
			t.Errorf("unexpected question type: %s", qType)
		}
	})

	t.Run("date input", func(t *testing.T) {
		rg := types.ItemComponent{
			Key:  "rg",
			Role: "responseGroup",
			Items: []types.ItemComponent{
				{Key: "1", Role: "dateInput"},
			},
		}
		ro, qType := extractResponses(&rg, testLang)
		if len(ro) != 1 {
			t.Error("shouldn't be empty")
		}
		if qType != QUESTION_TYPE_DATE_INPUT {
			t.Errorf("unexpected question type: %s", qType)
		}
	})

	t.Run("text input", func(t *testing.T) {
		rg := types.ItemComponent{
			Key:  "rg",
			Role: "responseGroup",
			Items: []types.ItemComponent{
				{Key: "1", Role: "input"},
			},
		}
		ro, qType := extractResponses(&rg, testLang)
		if len(ro) != 1 {
			t.Error("shouldn't be empty")
		}
		if qType != QUESTION_TYPE_TEXT_INPUT {
			t.Errorf("unexpected question type: %s", qType)
		}
	})

	t.Run("number input", func(t *testing.T) {
		rg := types.ItemComponent{
			Key:  "rg",
			Role: "responseGroup",
			Items: []types.ItemComponent{
				{Key: "1", Role: "numberInput"},
			},
		}
		ro, qType := extractResponses(&rg, testLang)
		if len(ro) != 1 {
			t.Error("shouldn't be empty")
		}
		if qType != QUESTION_TYPE_NUMBER_INPUT {
			t.Errorf("unexpected question type: %s", qType)
		}
	})

	t.Run("eq5d slider", func(t *testing.T) {
		rg := types.ItemComponent{
			Key:  "rg",
			Role: "responseGroup",
			Items: []types.ItemComponent{
				{Key: "1", Role: "eq5d-health-indicator"},
			},
		}
		ro, qType := extractResponses(&rg, testLang)
		if len(ro) != 1 {
			t.Error("shouldn't be empty")
		}
		if qType != QUESTION_TYPE_EQ5D_SLIDER {
			t.Errorf("unexpected question type: %s", qType)
		}
	})

	t.Run("numeric slider", func(t *testing.T) {
		rg := types.ItemComponent{
			Key:  "rg",
			Role: "responseGroup",
			Items: []types.ItemComponent{
				{Key: "1", Role: "sliderNumeric"},
			},
		}
		ro, qType := extractResponses(&rg, testLang)
		if len(ro) != 1 {
			t.Error("shouldn't be empty")
		}
		if qType != QUESTION_TYPE_NUMERIC_SLIDER {
			t.Errorf("unexpected question type: %s", qType)
		}
	})

	t.Run("dropdown group", func(t *testing.T) {
		rg := types.ItemComponent{
			Key:  "rg",
			Role: "responseGroup",
			Items: []types.ItemComponent{
				{Key: "ddg", Role: "dropDownGroup", Items: []types.ItemComponent{
					{Key: "1", Role: "option"},
					{Key: "2", Role: "option"},
					{Key: "3", Role: "option"},
				}},
			},
		}
		ro, qType := extractResponses(&rg, testLang)
		if len(ro) != 1 {
			t.Error("shouldn't be empty")
		}
		if qType != QUESTION_TYPE_DROPDOWN {
			t.Errorf("unexpected question type: %s", qType)
		}
	})

	t.Run("matrix", func(t *testing.T) {
		rg := types.ItemComponent{
			Key:  "rg",
			Role: "responseGroup",
			Items: []types.ItemComponent{
				{Key: "m", Role: "matrix", Items: []types.ItemComponent{
					{Key: "r1", Role: "responseRow", Items: []types.ItemComponent{
						{Key: "c1", Role: "label"},
						{Key: "c2", Role: "dropDownGroup", Items: []types.ItemComponent{
							{Key: "1", Role: "option"},
							{Key: "2", Role: "option"},
							{Key: "3", Role: "option"},
						}},
						{Key: "c3", Role: "check"},
					}},
					{Key: "r2", Role: "responseRow", Items: []types.ItemComponent{
						{Key: "c1", Role: "label"},
						{Key: "c2", Role: "dropDownGroup", Items: []types.ItemComponent{
							{Key: "1", Role: "option"},
							{Key: "2", Role: "option"},
							{Key: "3", Role: "option"},
						}},
						{Key: "c3", Role: "input"},
					}},
					{Key: "r3", Role: "radioRow", Items: []types.ItemComponent{
						{Key: "c1", Role: "label"},
						{Key: "c2", Role: "option"},
						{Key: "c3", Role: "option"},
					}},
				}},
			},
		}
		ro, qType := extractResponses(&rg, testLang)
		if len(ro) != 5 {
			t.Error("shouldn't be empty")
		}
		if qType != QUESTION_TYPE_MATRIX {
			t.Errorf("unexpected question type: %s", qType)
		}
	})
}

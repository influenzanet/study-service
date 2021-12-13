package exporter

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"testing"

	studyAPI "github.com/influenzanet/study-service/pkg/api"
	"github.com/influenzanet/study-service/pkg/types"
)

/*
"likertGroup"
"eq5d-health-indicator"
"multipleChoiceGroup"
"singleChoiceGroup"

"responseGroup"
*/

func TestResponseExporter(t *testing.T) {
	testLang := "en"
	questionOptionSep := "-"
	testSurveyDef := &studyAPI.SurveyItem{
		Key: "weekly",
		Items: []*studyAPI.SurveyItem{
			mockQuestion("weekly.Q1", testLang, "Title of Q1", mockSingleChoiceGroup(testLang, []MockOpionDef{
				{Key: "1", Role: "option", Label: "Yes"},
				{Key: "2", Role: "option", Label: "No"},
				{Key: "3", Role: "input", Label: "Other"},
			})),
			mockQuestion("weekly.Q2", testLang, "Title of Q2", mockMultipleChoiceGroup(testLang, []MockOpionDef{
				{Key: "1", Role: "option", Label: "Option 1"},
				{Key: "2", Role: "option", Label: "Option 2"},
				{Key: "3", Role: "input", Label: "Other"},
			})),
			{Key: "weekly.G1", Items: []*studyAPI.SurveyItem{
				mockQuestion("weekly.G1.Q1", testLang, "Title of Group 1's Q1", mockLikertGroup(testLang, []MockOpionDef{
					{Key: "cat1", Label: "Category 1"},
					{Key: "cat2", Label: "Category 2"},
				}, []string{
					"o1", "o2", "o3",
				})),
			}},
		},
	}

	t.Run("with with missing surveyDef", func(t *testing.T) {
		_, err := NewResponseExporter(nil, "en", true, questionOptionSep)
		if err == nil {
			t.Error("error expected")
			return
		}
		if err.Error() != "current survey definition not found" {
			t.Errorf("unexpected error message: %v", err)
			return
		}
	})

	t.Run("with with missing current", func(t *testing.T) {
		testSurvey := studyAPI.Survey{
			Id:      "test-id",
			Current: nil,
			History: []*studyAPI.SurveyVersion{},
		}

		_, err := NewResponseExporter(&testSurvey, "en", true, questionOptionSep)
		if err == nil {
			t.Error("error expected")
			return
		}
		if err.Error() != "current survey definition not found" {
			t.Errorf("unexpected error message: %v", err)
			return
		}
	})

	t.Run("with with one version", func(t *testing.T) {
		testSurvey := studyAPI.Survey{
			Id: "test-id",
			Current: &studyAPI.SurveyVersion{
				Published:        10,
				VersionId:        "1",
				SurveyDefinition: testSurveyDef,
			},
		}

		rp, err := NewResponseExporter(&testSurvey, "en", true, questionOptionSep)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if len(rp.surveyVersions) != 1 {
			t.Errorf("unexpected number of versions: %d", len(rp.surveyVersions))
			return
		}

		if len(rp.surveyVersions[0].Questions) != 3 {
			t.Errorf("unexpected number of versions: %d", len(rp.surveyVersions[0].Questions))
			return
		}
	})

	t.Run("with with multiple versions", func(t *testing.T) {
		testSurvey := studyAPI.Survey{
			Id: "test-id",
			Current: &studyAPI.SurveyVersion{
				Published:        10,
				VersionId:        "3",
				SurveyDefinition: testSurveyDef,
			},
			History: []*studyAPI.SurveyVersion{
				{
					Published:        2,
					Unpublished:      5,
					VersionId:        "1",
					SurveyDefinition: testSurveyDef,
				},
				{
					Published:        5,
					Unpublished:      10,
					VersionId:        "2",
					SurveyDefinition: testSurveyDef,
				},
			},
		}

		rp, err := NewResponseExporter(&testSurvey, "en", true, questionOptionSep)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if len(rp.surveyVersions) != 3 {
			t.Errorf("unexpected number of versions: %d", len(rp.surveyVersions))
			return
		}
	})
}

func readTestFileToBytes(t *testing.T, fileName string) []byte {
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Errorf("Failed to read test-file: %s - %v", fileName, err)
	}
	return content
}

func TestExportFormats(t *testing.T) {
	var testSurvey types.Survey
	json.Unmarshal(readTestFileToBytes(t, "./test_files/testSurveyDef.json"), &testSurvey)

	parser, err := NewResponseExporter(testSurvey.ToAPI(), "nl", true, "-")
	if err != nil {
		t.Errorf("unexpected error: %v", err.Error())
		return
	}

	t.Run("with no responses added yet", func(t *testing.T) {
		buf := new(bytes.Buffer)
		err := parser.GetResponsesCSV(buf, nil)
		if err == nil {
			t.Error("should produce error")
		}
	})

	var testResponses []types.SurveyResponse
	json.Unmarshal(readTestFileToBytes(t, "./test_files/testResponses.json"), &testResponses)

	for _, response := range testResponses {
		err = parser.AddResponse(response.ToAPI())
		if err != nil {
			t.Errorf("unexpected error: %v", err.Error())
			return
		}
	}

	wideCSV := string(readTestFileToBytes(t, "./test_files/export_wide.csv"))
	t.Run("Wide CSV", func(t *testing.T) {
		buf := new(bytes.Buffer)
		err := parser.GetResponsesCSV(buf, nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if buf.String() != wideCSV {
			t.Errorf("Unexpected output: %v", buf.String())
		}
	})

	longCSV := string(readTestFileToBytes(t, "./test_files/export_long.csv"))
	t.Run("Long CSV", func(t *testing.T) {
		buf := new(bytes.Buffer)
		err := parser.GetResponsesLongFormatCSV(buf, nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if buf.String() != longCSV {
			t.Errorf("Unexpected output: %v", buf.String())
		}
	})

	json := string(readTestFileToBytes(t, "./test_files/export.json"))
	t.Run("JSON", func(t *testing.T) {
		buf := new(bytes.Buffer)
		err := parser.GetResponsesJSON(buf, nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if buf.String() != json {
			t.Errorf("Unexpected output: %v", buf.String())
		}
	})
}

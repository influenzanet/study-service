package exporter

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/coneno/logger"
	"github.com/influenzanet/study-service/pkg/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestResponseExporter(t *testing.T) {
	logger.SetLevel(logger.LEVEL_ERROR)
	testLang := "en"
	questionOptionSep := "-"
	testSurveyDef := &types.SurveyItem{
		Key: "weekly",
		Items: []types.SurveyItem{
			*mockQuestion("weekly.Q1", testLang, "Title of Q1", mockSingleChoiceGroup(testLang, []MockOpionDef{
				{Key: "1", Role: "option", Label: "Yes"},
				{Key: "2", Role: "option", Label: "No"},
				{Key: "3", Role: "input", Label: "Other"},
			})),
			*mockQuestion("weekly.Q2", testLang, "Title of Q2", mockMultipleChoiceGroup(testLang, []MockOpionDef{
				{Key: "1", Role: "option", Label: "Option 1"},
				{Key: "2", Role: "option", Label: "Option 2"},
				{Key: "3", Role: "input", Label: "Other"},
			})),
			{Key: "weekly.G1", Items: []types.SurveyItem{
				*mockQuestion("weekly.G1.Q1", testLang, "Title of Group 1's Q1", mockLikertGroup(testLang, []MockOpionDef{
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
		if err.Error() != "survey definition history not found" {
			t.Errorf("unexpected error message: %v", err)
			return
		}
	})

	// Can't miss anymore
	// t.Run("with with missing current", func(t *testing.T) {
	// 	_id, idErr := primitive.ObjectIDFromHex("5ed7497024c0797b0a41b1ca")
	// 	if idErr != nil {
	// 		t.Errorf("unexpected error message: %v", idErr)
	// 		return
	// 	}
	// 	testSurvey := types.Survey{
	// 		ID:      _id,
	// 		Current: types.SurveyVersion{},
	// 		History: []types.SurveyVersion{},
	// 	}

	// 	_, err := NewResponseExporter(&testSurvey, "en", true, questionOptionSep)
	// 	if err == nil {
	// 		t.Error("error expected")
	// 		return
	// 	}
	// 	if err.Error() != "survey definition history not found" {
	// 		t.Errorf("unexpected error message: %v", err)
	// 		return
	// 	}
	// })

	t.Run("with with one version", func(t *testing.T) {
		_id, idErr := primitive.ObjectIDFromHex("5ed7497024c0797b0a41b1ca")
		if idErr != nil {
			t.Errorf("unexpected error message: %v", idErr)
			return
		}
		testSurvey := types.Survey{
			ID:               _id,
			Published:        10,
			VersionID:        "1",
			SurveyDefinition: *testSurveyDef,
		}

		rp, err := NewResponseExporter([]*types.Survey{&testSurvey}, "en", true, questionOptionSep)
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
		_id, idErr := primitive.ObjectIDFromHex("5ed7497024c0797b0a41b1ca")
		if idErr != nil {
			t.Errorf("unexpected error message: %v", idErr)
			return
		}
		v1 := &types.Survey{
			Published:        2,
			Unpublished:      5,
			VersionID:        "1",
			SurveyDefinition: *testSurveyDef,
		}
		v2 := &types.Survey{
			Published:        5,
			Unpublished:      10,
			VersionID:        "2",
			SurveyDefinition: *testSurveyDef,
		}
		v3 := &types.Survey{
			ID:               _id,
			Published:        10,
			VersionID:        "3",
			SurveyDefinition: *testSurveyDef,
		}

		rp, err := NewResponseExporter([]*types.Survey{
			v1, v2, v3,
		}, "en", true, questionOptionSep)
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
	content, err := os.ReadFile(fileName)
	if err != nil {
		t.Errorf("Failed to read test-file: %s - %v", fileName, err)
	}
	return content
}

func TestExportFormats(t *testing.T) {
	var testSurveyHistory types.SurveyVersionsJSON
	json.Unmarshal(readTestFileToBytes(t, "./test_files/testSurveyDef.json"), &testSurveyHistory)

	parser, err := NewResponseExporter(testSurveyHistory.SurveyVersions, "nl", true, "-")
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
		err = parser.AddResponse(&response)
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
			// t.Errorf("Unexpected output: %v", buf.String())
			t.Errorf("Unexpected output")
			writeBytesToFile(buf.Bytes(), "./test_files/error/export_wide.csv")
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
			// t.Errorf("Unexpected output: %v", buf.String())
			t.Errorf("Unexpected output")
			writeBytesToFile(buf.Bytes(), "./test_files/error/export_long.csv")
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
			// t.Errorf("Unexpected output: %v", buf.String())
			t.Errorf("Unexpected output")
			writeBytesToFile(buf.Bytes(), "./test_files/error/export.json")
		}
	})
}

func TestExportFormatsExcludeFilter(t *testing.T) {
	var testSurveyHistory types.SurveyVersionsJSON
	json.Unmarshal(readTestFileToBytes(t, "./test_files/testSurveyDef.json"), &testSurveyHistory)

	parser, err := NewResponseExporterWithExcludeFilter(testSurveyHistory.SurveyVersions, "nl", true, "-", []string{"weekly.HS.Q11", "weekly.HS.contact.Q7"})
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
		err = parser.AddResponse(&response)
		if err != nil {
			t.Errorf("unexpected error: %v", err.Error())
			return
		}
	}

	wideCSV := string(readTestFileToBytes(t, "./test_files/excludeFilter/export_wide.csv"))
	t.Run("Wide CSV", func(t *testing.T) {
		buf := new(bytes.Buffer)
		err := parser.GetResponsesCSV(buf, nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if buf.String() != wideCSV {
			// t.Errorf("Unexpected output: %v", buf.String())
			t.Errorf("Unexpected output")
			writeBytesToFile(buf.Bytes(), "./test_files/error/export_wide.csv")
		}
	})

	longCSV := string(readTestFileToBytes(t, "./test_files/excludeFilter/export_long.csv"))
	t.Run("Long CSV", func(t *testing.T) {
		buf := new(bytes.Buffer)
		err := parser.GetResponsesLongFormatCSV(buf, nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if buf.String() != longCSV {
			// t.Errorf("Unexpected output: %v", buf.String())
			t.Errorf("Unexpected output")
			writeBytesToFile(buf.Bytes(), "./test_files/error/export_long.csv")
		}
	})

	json := string(readTestFileToBytes(t, "./test_files/excludeFilter/export.json"))
	t.Run("JSON", func(t *testing.T) {
		buf := new(bytes.Buffer)
		err := parser.GetResponsesJSON(buf, nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if buf.String() != json {
			// t.Errorf("Unexpected output: %v", buf.String())
			t.Errorf("Unexpected output")
			writeBytesToFile(buf.Bytes(), "./test_files/error/export.json")
		}
	})
}

func TestExportFormatsIncludeFilter(t *testing.T) {
	var testSurveyHistory types.SurveyVersionsJSON
	json.Unmarshal(readTestFileToBytes(t, "./test_files/testSurveyDef.json"), &testSurveyHistory)

	parser, err := NewResponseExporterWithIncludeFilter(testSurveyHistory.SurveyVersions, "nl", true, "-", []string{"weekly.HS.Q11", "weekly.HS.contact.Q7"})
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
		err = parser.AddResponse(&response)
		if err != nil {
			t.Errorf("unexpected error: %v", err.Error())
			return
		}
	}

	wideCSV := string(readTestFileToBytes(t, "./test_files/includeFilter/export_wide.csv"))
	t.Run("Wide CSV", func(t *testing.T) {
		buf := new(bytes.Buffer)
		err := parser.GetResponsesCSV(buf, nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if buf.String() != wideCSV {
			// t.Errorf("Unexpected output: %v", buf.String())
			t.Errorf("Unexpected output")
			writeBytesToFile(buf.Bytes(), "./test_files/error/export_wide.csv")
		}
	})

	longCSV := string(readTestFileToBytes(t, "./test_files/includeFilter/export_long.csv"))
	t.Run("Long CSV", func(t *testing.T) {
		buf := new(bytes.Buffer)
		err := parser.GetResponsesLongFormatCSV(buf, nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if buf.String() != longCSV {
			t.Errorf("Unexpected output: %v", buf.String())
			// t.Errorf("Unexpected output")
			writeBytesToFile(buf.Bytes(), "./test_files/error/export_long.csv")
		}
	})

	json := string(readTestFileToBytes(t, "./test_files/includeFilter/export.json"))
	t.Run("JSON", func(t *testing.T) {
		buf := new(bytes.Buffer)
		err := parser.GetResponsesJSON(buf, nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if buf.String() != json {
			// t.Errorf("Unexpected output: %v", buf.String())
			t.Errorf("Unexpected output")
			writeBytesToFile(buf.Bytes(), "./test_files/error/export.json")
		}
	})
}

func TestExportFormatsWithMetaTimestamps(t *testing.T) {
	var testSurveyHistory types.SurveyVersionsJSON
	json.Unmarshal(readTestFileToBytes(t, "./test_files/testSurveyDef.json"), &testSurveyHistory)

	parser, err := NewResponseExporter(testSurveyHistory.SurveyVersions, "nl", true, "-")
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
		err = parser.AddResponse(&response)
		if err != nil {
			t.Errorf("unexpected error: %v", err.Error())
			return
		}
	}

	wideCSV := string(readTestFileToBytes(t, "./test_files/metaTimestamps/export_wide.csv"))
	t.Run("Wide CSV", func(t *testing.T) {
		buf := new(bytes.Buffer)
		err := parser.GetResponsesCSV(buf, &IncludeMeta{Postion: false, InitTimes: true, DisplayedTimes: true, ResponsedTimes: true})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if buf.String() != wideCSV {
			// t.Errorf("Unexpected output: %v", buf.String())
			t.Errorf("Unexpected output")
			writeBytesToFile(buf.Bytes(), "./test_files/error/export_wide.csv")
		}
	})

	longCSV := string(readTestFileToBytes(t, "./test_files/metaTimestamps/export_long.csv"))
	t.Run("Long CSV", func(t *testing.T) {
		buf := new(bytes.Buffer)
		err := parser.GetResponsesLongFormatCSV(buf, &IncludeMeta{Postion: false, InitTimes: true, DisplayedTimes: true, ResponsedTimes: true})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if buf.String() != longCSV {
			// t.Errorf("Unexpected output: %v", buf.String())
			t.Errorf("Unexpected output")
			writeBytesToFile(buf.Bytes(), "./test_files/error/export_long.csv")
		}
	})

	json := string(readTestFileToBytes(t, "./test_files/metaTimestamps/export.json"))
	t.Run("JSON", func(t *testing.T) {
		buf := new(bytes.Buffer)
		err := parser.GetResponsesJSON(buf, &IncludeMeta{Postion: false, InitTimes: true, DisplayedTimes: true, ResponsedTimes: true})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if buf.String() != json {
			// t.Errorf("Unexpected output: %v", buf.String())
			t.Errorf("Unexpected output")
			writeBytesToFile(buf.Bytes(), "./test_files/error/export.json")
		}
	})
}

func TestExportFormatsQuestionTypes(t *testing.T) {
	var testSurveyHistory types.SurveyVersionsJSON
	json.Unmarshal(readTestFileToBytes(t, "./test_files/questionTypes/surveyDef.json"), &testSurveyHistory)

	parser, err := NewResponseExporter(testSurveyHistory.SurveyVersions, "nl", true, "-")
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
	json.Unmarshal(readTestFileToBytes(t, "./test_files/questionTypes/responses.json"), &testResponses)

	for _, response := range testResponses {
		err = parser.AddResponse(&response)
		if err != nil {
			t.Errorf("unexpected error: %v", err.Error())
			return
		}
	}

	wideCSV := string(readTestFileToBytes(t, "./test_files/questionTypes/export_wide.csv"))
	t.Run("Wide CSV", func(t *testing.T) {
		buf := new(bytes.Buffer)
		err := parser.GetResponsesCSV(buf, nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if buf.String() != wideCSV {
			//t.Errorf("Unexpected output: %v", buf.String())
			t.Errorf("Unexpected output")
			writeBytesToFile(buf.Bytes(), "./test_files/error/export_wide_qt.csv")
		}
	})

	longCSV := string(readTestFileToBytes(t, "./test_files/questionTypes/export_long.csv"))
	t.Run("Long CSV", func(t *testing.T) {
		buf := new(bytes.Buffer)
		err := parser.GetResponsesLongFormatCSV(buf, nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if buf.String() != longCSV {
			// t.Errorf("Unexpected output: %v", buf.String())
			t.Errorf("Unexpected output")
			writeBytesToFile(buf.Bytes(), "./test_files/error/export_long_qt.csv")
		}
	})

	json := string(readTestFileToBytes(t, "./test_files/questionTypes/export.json"))
	t.Run("JSON", func(t *testing.T) {
		buf := new(bytes.Buffer)
		err := parser.GetResponsesJSON(buf, nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if buf.String() != json {
			// t.Errorf("Unexpected output: %v", buf.String())
			t.Errorf("Unexpected output")
			writeBytesToFile(buf.Bytes(), "./test_files/error/export_qt.json")
		}
	})
}

func writeBytesToFile(bytes []byte, fileName string) {
	f, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	f.Write(bytes)
}

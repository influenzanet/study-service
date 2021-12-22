package exporter

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/coneno/logger"
	"github.com/influenzanet/study-service/pkg/types"
)

func findSurveyVersion(versionID string, submittedAt int64, versions []SurveyVersionPreview) (sv SurveyVersionPreview, err error) {
	if versionID == "" {
		return findVersionBasedOnTimestamp(submittedAt, versions)
	} else {
		sv, err = findVersionBasedOnVersionID(versionID, versions)
		if err != nil {
			return findVersionBasedOnTimestamp(submittedAt, versions)
		}
	}
	return sv, nil
}

func findVersionBasedOnTimestamp(submittedAt int64, versions []SurveyVersionPreview) (sv SurveyVersionPreview, err error) {
	for _, v := range versions {
		if v.Unpublished == 0 {
			if v.Published <= submittedAt {
				return v, nil
			}
		} else {
			if v.Published <= submittedAt && v.Unpublished > submittedAt {
				return v, nil
			}
		}
	}
	return sv, fmt.Errorf("no survey version found: %d", submittedAt)
}

func findVersionBasedOnVersionID(versionID string, versions []SurveyVersionPreview) (sv SurveyVersionPreview, err error) {
	for _, v := range versions {
		if v.VersionID == versionID {
			return v, nil
		}
	}
	return sv, errors.New("no survey version found")
}

func timestampsToStr(ts []int64) string {
	bytes, err := json.Marshal(ts)
	if err != nil {
		return err.Error()
	}

	return string(bytes)
}

func findResponse(responses []types.SurveyItemResponse, key string) *types.SurveyItemResponse {
	for _, r := range responses {
		if r.Key == key {
			return &r
		}
	}
	return nil
}

func getResponseColumns(question SurveyQuestion, response *types.SurveyItemResponse, questionOptionSep string) map[string]interface{} {
	switch question.QuestionType {
	case QUESTION_TYPE_SINGLE_CHOICE:
		return processResponseForSingleChoice(question, response, questionOptionSep)
	case QUESTION_TYPE_DROPDOWN:
		return processResponseForSingleChoice(question, response, questionOptionSep)
	case QUESTION_TYPE_LIKERT:
		return processResponseForSingleChoice(question, response, questionOptionSep)
	case QUESTION_TYPE_LIKERT_GROUP:
		return handleSingleChoiceGroupList(question.ID, question.Responses, response, questionOptionSep)
	case QUESTION_TYPE_RESPONSIVE_SINGLE_CHOICE_ARRAY:
		return processResponseForSingleChoice(question, response, questionOptionSep)
	case QUESTION_TYPE_RESPONSIVE_BIPOLAR_LIKERT_ARRAY:
		return processResponseForSingleChoice(question, response, questionOptionSep)
	case QUESTION_TYPE_MULTIPLE_CHOICE:
		return processResponseForMultipleChoice(question, response, questionOptionSep)
	case QUESTION_TYPE_TEXT_INPUT:
		return processResponseForInputs(question, response, questionOptionSep)
	case QUESTION_TYPE_DATE_INPUT:
		return processResponseForInputs(question, response, questionOptionSep)
	case QUESTION_TYPE_NUMBER_INPUT:
		return processResponseForInputs(question, response, questionOptionSep)
	case QUESTION_TYPE_NUMERIC_SLIDER:
		return processResponseForInputs(question, response, questionOptionSep)
	case QUESTION_TYPE_EQ5D_SLIDER:
		return processResponseForInputs(question, response, questionOptionSep)
	case QUESTION_TYPE_MATRIX:
		return processResponseForMatrix(question, response, questionOptionSep)
	case QUESTION_TYPE_UNKNOWN:
		return processResponseForUnknown(question, response, questionOptionSep)
	default:
		return map[string]interface{}{}
	}
}

func processResponseForSingleChoice(question SurveyQuestion, response *types.SurveyItemResponse, questionOptionSep string) map[string]interface{} {
	var responseCols map[string]interface{}

	if len(question.Responses) == 1 {
		rSlot := question.Responses[0]
		responseCols = handleSimpleSingleChoiceGroup(question.ID, rSlot, response, questionOptionSep)

	} else {
		responseCols = handleSingleChoiceGroupList(question.ID, question.Responses, response, questionOptionSep)
	}
	return responseCols
}

func handleSimpleSingleChoiceGroup(questionKey string, responseSlotDef ResponseDef, response *types.SurveyItemResponse, questionOptionSep string) map[string]interface{} {
	responseCols := map[string]interface{}{}

	// Prepare columns:
	responseCols[questionKey] = ""

	for _, option := range responseSlotDef.Options {
		if option.OptionType != OPTION_TYPE_RADIO &&
			option.OptionType != OPTION_TYPE_DROPDOWN_OPTION {
			responseCols[questionKey+questionOptionSep+option.ID] = ""
		}
	}

	// Find responses
	rGroup := retrieveResponseItem(response, RESPONSE_ROOT_KEY+"."+responseSlotDef.ID)
	if rGroup != nil {
		if len(rGroup.Items) != 1 {
			logger.Debug.Printf("unexpected response group for question %s: %v", questionKey, rGroup)
		} else {
			selection := rGroup.Items[0]
			responseCols[questionKey] = selection.Key

			valueKey := questionKey + questionOptionSep + selection.Key
			if _, hasKey := responseCols[valueKey]; hasKey {
				responseCols[valueKey] = selection.Value
			}
		}
	}
	return responseCols
}

func handleSingleChoiceGroupList(questionKey string, responseSlotDefs []ResponseDef, response *types.SurveyItemResponse, questionOptionSep string) map[string]interface{} {
	responseCols := map[string]interface{}{}

	// Prepare columns:
	for _, rSlot := range responseSlotDefs {
		responseCols[questionKey+questionOptionSep+rSlot.ID] = ""
		for _, option := range rSlot.Options {
			if option.OptionType != OPTION_TYPE_RADIO &&
				option.OptionType != OPTION_TYPE_DROPDOWN_OPTION {
				responseCols[questionKey+questionOptionSep+rSlot.ID+"."+option.ID] = ""
			}
		}
	}

	// Find responses:
	for _, rSlot := range responseSlotDefs {
		rGroup := retrieveResponseItemByShortKey(response, rSlot.ID)
		if rGroup == nil {
			continue
		} else if len(rGroup.Items) != 1 {
			logger.Debug.Printf("unexpected response group for question %s: %v", questionKey, rGroup)
			continue
		}

		selection := rGroup.Items[0]
		responseCols[questionKey+questionOptionSep+rSlot.ID] = selection.Key

		valueKey := questionKey + questionOptionSep + rSlot.ID + "." + selection.Key
		if _, hasKey := responseCols[valueKey]; hasKey {
			responseCols[valueKey] = selection.Value
		}
	}
	return responseCols
}

func processResponseForMultipleChoice(question SurveyQuestion, response *types.SurveyItemResponse, questionOptionSep string) map[string]interface{} {
	var responseCols map[string]interface{}

	if len(question.Responses) == 1 {
		rSlot := question.Responses[0]
		responseCols = handleSimpleMultipleChoiceGroup(question.ID, rSlot, response, questionOptionSep)

	} else {
		responseCols = handleMultipleChoiceGroupList(question.ID, question.Responses, response, questionOptionSep)
	}
	return responseCols
}

func handleSimpleMultipleChoiceGroup(questionKey string, responseSlotDef ResponseDef, response *types.SurveyItemResponse, questionOptionSep string) map[string]interface{} {
	responseCols := map[string]interface{}{}

	// Find responses
	rGroup := retrieveResponseItem(response, RESPONSE_ROOT_KEY+"."+responseSlotDef.ID)
	if rGroup != nil {
		if len(rGroup.Items) > 0 {
			for _, option := range responseSlotDef.Options {
				responseCols[questionKey+questionOptionSep+option.ID] = FALSE_VALUE
				if option.OptionType != OPTION_TYPE_CHECKBOX {
					responseCols[questionKey+questionOptionSep+option.ID+questionOptionSep+OPEN_FIELD_COL_SUFFIX] = ""
				}
			}

			for _, item := range rGroup.Items {
				responseCols[questionKey+questionOptionSep+item.Key] = TRUE_VALUE

				valueKey := questionKey + questionOptionSep + item.Key + questionOptionSep + OPEN_FIELD_COL_SUFFIX
				if _, hasKey := responseCols[valueKey]; hasKey {
					responseCols[valueKey] = item.Value
				}
			}
		}
	} else {
		for _, option := range responseSlotDef.Options {
			responseCols[questionKey+questionOptionSep+option.ID] = ""
			if option.OptionType != OPTION_TYPE_CHECKBOX {
				responseCols[questionKey+questionOptionSep+option.ID+questionOptionSep+OPEN_FIELD_COL_SUFFIX] = ""
			}
		}

	}
	return responseCols
}

func handleMultipleChoiceGroupList(questionKey string, responseSlotDefs []ResponseDef, response *types.SurveyItemResponse, questionOptionSep string) map[string]interface{} {
	responseCols := map[string]interface{}{}

	// Prepare columns:
	for _, rSlot := range responseSlotDefs {
		// Find responses
		rGroup := retrieveResponseItem(response, RESPONSE_ROOT_KEY+"."+rSlot.ID)
		slotKeyPrefix := questionKey + questionOptionSep + rSlot.ID + "."
		if rGroup != nil {
			if len(rGroup.Items) > 0 {
				for _, option := range rSlot.Options {
					responseCols[slotKeyPrefix+option.ID] = FALSE_VALUE
					if option.OptionType != OPTION_TYPE_CHECKBOX {
						responseCols[slotKeyPrefix+option.ID+questionOptionSep+OPEN_FIELD_COL_SUFFIX] = ""
					}
				}

				for _, item := range rGroup.Items {
					responseCols[slotKeyPrefix+item.Key] = TRUE_VALUE

					valueKey := slotKeyPrefix + item.Key + questionOptionSep + OPEN_FIELD_COL_SUFFIX
					if _, hasKey := responseCols[valueKey]; hasKey {
						responseCols[valueKey] = item.Value
					}
				}
			}
		} else {
			for _, option := range rSlot.Options {
				responseCols[slotKeyPrefix+option.ID] = ""
				if option.OptionType != OPTION_TYPE_CHECKBOX {
					responseCols[slotKeyPrefix+option.ID+questionOptionSep+OPEN_FIELD_COL_SUFFIX] = ""
				}
			}

		}
	}

	return responseCols
}

func processResponseForInputs(question SurveyQuestion, response *types.SurveyItemResponse, questionOptionSep string) map[string]interface{} {
	var responseCols map[string]interface{}

	if len(question.Responses) == 1 {
		rSlot := question.Responses[0]
		responseCols = handleSimpleInput(question.ID, rSlot, response, questionOptionSep)

	} else {
		responseCols = handleInputList(question.ID, question.Responses, response, questionOptionSep)
	}

	return responseCols
}

func handleSimpleInput(questionKey string, responseSlotDef ResponseDef, response *types.SurveyItemResponse, questionOptionSep string) map[string]interface{} {
	responseCols := map[string]interface{}{}
	responseCols[questionKey] = ""

	// Find responses
	rValue := retrieveResponseItem(response, RESPONSE_ROOT_KEY+"."+responseSlotDef.ID)
	if rValue != nil {
		responseCols[questionKey] = rValue.Value
	}
	return responseCols
}

func handleInputList(questionKey string, responseSlotDefs []ResponseDef, response *types.SurveyItemResponse, questionOptionSep string) map[string]interface{} {
	responseCols := map[string]interface{}{}

	for _, rSlot := range responseSlotDefs {
		// Prepare columns:
		slotKey := questionKey + questionOptionSep + rSlot.ID
		responseCols[slotKey] = ""

		// Find responses
		rValue := retrieveResponseItem(response, RESPONSE_ROOT_KEY+"."+rSlot.ID)
		if rValue != nil {
			responseCols[slotKey] = rValue.Value
		}
	}

	return responseCols
}

func processResponseForMatrix(question SurveyQuestion, response *types.SurveyItemResponse, questionOptionSep string) map[string]interface{} {
	responseCols := map[string]interface{}{}

	for _, rSlot := range question.Responses {
		// Prepare columns:
		slotKey := question.ID + questionOptionSep + rSlot.ID

		if rSlot.ResponseType == QUESTION_TYPE_MATRIX_RADIO_ROW {
			rGroup := retrieveResponseItem(response, RESPONSE_ROOT_KEY+"."+rSlot.ID)
			responseCols[slotKey] = ""
			if rGroup != nil {
				if len(rGroup.Items) != 1 {
					logger.Debug.Printf("unexpected response group for question %s: %v", question.ID, rGroup)
				} else {
					selection := rGroup.Items[0]
					responseCols[slotKey] = selection.Key
				}
			}
		} else {
			rGroup := retrieveResponseItem(response, RESPONSE_ROOT_KEY+"."+rSlot.ID)
			responseCols[slotKey] = ""
			if rGroup != nil {
				if len(rGroup.Items) != 1 {
					logger.Debug.Printf("unexpected response group for question %s: %v", question.ID, rGroup)
				} else {
					selection := rGroup.Items[0]
					value := selection.Key
					if selection.Value != "" {
						value = selection.Value
					}
					responseCols[slotKey] = value
				}
			}
		}
	}
	return responseCols
}

func processResponseForUnknown(question SurveyQuestion, response *types.SurveyItemResponse, questionOptionSep string) map[string]interface{} {
	responseCols := map[string]interface{}{}

	for _, rSlot := range question.Responses {
		slotKey := question.ID + questionOptionSep + rSlot.ID
		responseCols[slotKey] = ""
		rGroup := retrieveResponseItem(response, RESPONSE_ROOT_KEY+"."+rSlot.ID)
		if rGroup != nil {
			responseCols[slotKey] = rGroup
		}
	}
	return responseCols
}

func retrieveResponseItem(response *types.SurveyItemResponse, fullKey string) *types.ResponseItem {
	if response == nil || response.Response == nil {
		return nil
	}
	keyParts := strings.Split(fullKey, ".")

	var result *types.ResponseItem
	for _, key := range keyParts {
		if result == nil {
			if key != response.Response.Key {
				return nil
			}
			result = response.Response
			continue
		}
		found := false
		for _, item := range result.Items {
			if item.Key == key {
				result = &item
				found = true
				break
			}
		}
		if !found {
			return nil
		}
	}
	return result
}

func retrieveResponseItemByShortKey(response *types.SurveyItemResponse, shortKey string) *types.ResponseItem {
	if response == nil || response.Response == nil {
		return nil
	}

	var result *types.ResponseItem
	if response.Response.Key == shortKey {
		return response.Response
	}

	result = response.Response

	for _, item := range result.Items {
		if item.Key == shortKey {
			return &item
		}
	}

	for _, item := range result.Items {
		res := retrieveResponseItemByShortKey(&types.SurveyItemResponse{
			Response: &item,
		}, shortKey)
		if res != nil {
			return res
		}
	}
	return nil
}

func responseColToString(responseCol interface{}) string {
	var str string
	switch colValue := responseCol.(type) {
	case string:
		str = colValue
	case *types.ResponseItem:
		jsonBytes, err := json.Marshal(colValue)
		if err != nil {
			logger.Debug.Printf("error while parsing response column: %v", err)
			return err.Error()
		}
		str = string(jsonBytes)
	}
	return str
}

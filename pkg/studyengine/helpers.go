package studyengine

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/influenzanet/study-service/pkg/types"
)

// Method to find survey item response in the array of responses
func findSurveyItemResponse(responses []types.SurveyItemResponse, key string) (responseOfInterest *types.SurveyItemResponse, err error) {
	for _, response := range responses {
		if response.Key == key {
			return &response, nil
		}
	}
	return nil, errors.New("item not found")
}

// Method to retrive one level of the nested response object
func findResponseObject(surveyItem *types.SurveyItemResponse, responseKey string) (responseItem *types.ResponseItem, err error) {
	if surveyItem == nil {
		return responseItem, errors.New("missing survey item")
	}
	if surveyItem.Response == nil {
		return responseItem, errors.New("missing survey item response")
	}
	for i, k := range strings.Split(responseKey, ".") {
		if i == 0 {
			if surveyItem.Response.Key != k {
				// item not found:
				return responseItem, errors.New("response object is not found")
			}
			responseItem = surveyItem.Response
			continue
		}

		found := false
		for _, item := range responseItem.Items {
			if item.Key == k {
				found = true
				responseItem = &item
				break
			}
		}
		if !found {
			// item not found:
			return responseItem, errors.New("response object is not found")
		}
	}
	return responseItem, nil
}

func getExternalServicesConfigByName(serviceConfigs []types.ExternalService, name string) (types.ExternalService, error) {
	for _, item := range serviceConfigs {
		if item.Name == name {
			return item, nil
		}
	}
	return types.ExternalService{}, fmt.Errorf("no external service config found with name: %s", name)
}

type ExternalEventPayload struct {
	ParticipantState types.ParticipantState `json:"participantState"`
	EventType        string                 `json:"eventType"`
	StudyKey         string                 `json:"studyKey"`
	InstanceID       string                 `json:"instanceID"`
	Response         types.SurveyResponse   `json:"surveyResponses"`
}

func runHTTPcall(url string, APIkey string, payload ExternalEventPayload) (map[string]interface{}, error) {
	json_data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// TODO: send API key through header
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(json_data))
	if err != nil {
		return nil, err
	}

	var res map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

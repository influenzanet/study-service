package studyengine

import (
	"errors"
	"fmt"
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

package studyengine

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/coneno/logger"
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
				responseItem = item
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

type ClientConfig struct {
	APIKey     string
	mTLSConfig *types.MutualTLSConfig
	Timeout    time.Duration
}

func runHTTPcall(url string, payload ExternalEventPayload, config ClientConfig) (map[string]interface{}, error) {
	json_data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	transport, err := getTransportWithMTLSConfig(config.mTLSConfig)
	if err != nil {
		logger.Error.Printf("unexpected error: %v", err)
	}

	client := &http.Client{
		Timeout: config.Timeout,
	}
	if transport != nil {
		client.Transport = transport
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(json_data))
	if err != nil {
		logger.Error.Printf("unexpected error: %v", err)
		return nil, err
	}
	if config.APIKey != "" {
		req.Header.Set("Api-Key", config.APIKey)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.Debug.Printf("unexpected error: %v", err)
		return nil, err
	}

	var res map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		logger.Debug.Printf("unexpected error: %v", err)
		return nil, err
	}
	return res, nil
}

func getTransportWithMTLSConfig(mTLSConfig *types.MutualTLSConfig) (*http.Transport, error) {
	if mTLSConfig == nil {
		return nil, nil
	}

	// Load client certificate and key
	cert, err := tls.LoadX509KeyPair(mTLSConfig.CertFile, mTLSConfig.KeyFile)
	if err != nil {
		panic(err)
	}

	// Load CA certificate
	caCert, err := os.ReadFile(mTLSConfig.CAFile)
	if err != nil {
		panic(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Create TLS client configuration
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}

	return &http.Transport{
		TLSClientConfig: tlsConfig,
	}, nil
}

package studydb

import (
	"testing"

	"github.com/influenzanet/study-service/pkg/types"
)

func TestDbAddAndReplaceConfidentialResponse(t *testing.T) {
	testStudy := "testosaveconfidential"
	testResp := types.SurveyResponse{
		Key:           "test",
		ParticipantID: "test",
	}
	t.Run("add response", func(t *testing.T) {
		_, err := testDBService.AddConfidentialResponse(testInstanceID, testStudy, testResp)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
	})

	t.Run("replace response", func(t *testing.T) {
		err := testDBService.ReplaceConfidentialResponse(testInstanceID, testStudy, testResp)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
	})

	t.Run("replace response", func(t *testing.T) {
		testResp.Key = "test2"
		err := testDBService.ReplaceConfidentialResponse(testInstanceID, testStudy, testResp)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
	})

	t.Run("save response", func(t *testing.T) {
		testResp.Key = "test3"
		err := testDBService.ReplaceConfidentialResponse(testInstanceID, testStudy, testResp)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
	})
}

func TestDbDeleteConfidentialResponse(t *testing.T) {
	testStudy := "testosaveconfidential"
	testParticipantID := "test"
	t.Run("with key", func(t *testing.T) {
		c, err := testDBService.DeleteConfidentialResponses(testInstanceID, testStudy, testParticipantID, "test")
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if c != 1 {
			t.Errorf("unexpected number of deletions: %d", c)
		}
	})

	t.Run("without key", func(t *testing.T) {
		c, err := testDBService.DeleteConfidentialResponses(testInstanceID, testStudy, testParticipantID, "")
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if c != 2 {
			t.Errorf("unexpected number of deletions: %d", c)
		}
	})
}

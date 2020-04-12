package main

import (
	"testing"
	"time"

	"github.com/influenzanet/study-service/models"
)

func TestDbParticipantStateTest(t *testing.T) {
	testStudyKey := "teststudy234234"

	testPState := models.ParticipantState{
		ParticipantID: "testPID0990",
		StudyStatus:   "active",
		Flags: map[string]string{
			"testKey": "testValue",
		},
		LastSubmissions: map[string]int64{
			"testSurveyKey": time.Now().Unix(),
		},
	}

	t.Run("Testing find participant state, when not existing", func(t *testing.T) {
		_, err := findParticipantStateDB(testInstanceID, testStudyKey, testPState.ParticipantID)
		if err == nil {
			t.Error("should return an error")
		}
	})

	t.Run("Testing update participant state, when not existing", func(t *testing.T) {
		pState, err := saveParticipantStateDB(testInstanceID, testStudyKey, testPState)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if pState.ParticipantID != testPState.ParticipantID {
			t.Errorf("unexpected participantID -> have: %s, want: %s ", pState.ParticipantID, testPState.ParticipantID)
		}
	})

	t.Run("Testing find participant state, when existing", func(t *testing.T) {
		pState, err := findParticipantStateDB(testInstanceID, testStudyKey, testPState.ParticipantID)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if pState.ParticipantID != testPState.ParticipantID {
			t.Errorf("unexpected participantID -> have: %s, want: %s ", pState.ParticipantID, testPState.ParticipantID)
		}
	})

	t.Run("Testing update participant state, when existing", func(t *testing.T) {
		testPState.StudyStatus = "paused"
		pState, err := saveParticipantStateDB(testInstanceID, testStudyKey, testPState)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if pState.StudyStatus != testPState.StudyStatus {
			t.Errorf("unexpected participantID -> have: %s, want: %s ", pState.StudyStatus, testPState.StudyStatus)
		}

		pState, err = findParticipantStateDB(testInstanceID, testStudyKey, testPState.ParticipantID)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if pState.StudyStatus != testPState.StudyStatus {
			t.Errorf("unexpected participantID -> have: %s, want: %s ", pState.StudyStatus, testPState.StudyStatus)
		}
	})
}

func TestDbFindParticipantsByStatusTest(t *testing.T) {
	testStudyKey := "teststudy_findbystatus"

	pStates := []models.ParticipantState{
		{
			ParticipantID: "1",
			StudyStatus:   "active",
		},
		{
			ParticipantID: "2",
			StudyStatus:   "active",
		},
		{
			ParticipantID: "3",
			StudyStatus:   "inactive",
		},
		{
			ParticipantID: "4",
			StudyStatus:   "inactive",
		},
	}

	for _, ps := range pStates {
		_, err := saveParticipantStateDB(testInstanceID, testStudyKey, ps)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
	}

	t.Run("Finding active status", func(t *testing.T) {
		participants, err := findParticipantsByStudyStatusDB(testInstanceID, testStudyKey, "active", false)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(participants) != 2 {
			t.Errorf("unexpected number of participants found: %d, %d (have, want)", len(participants), 2)
		}
	})

	t.Run("Finding inactive status ", func(t *testing.T) {
		participants, err := findParticipantsByStudyStatusDB(testInstanceID, testStudyKey, "inactive", true)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(participants) != 2 {
			t.Errorf("unexpected number of participants found: %d, %d (have, want)", len(participants), 2)
		}
	})

	t.Run("Findign not existing status", func(t *testing.T) {
		participants, err := findParticipantsByStudyStatusDB(testInstanceID, testStudyKey, "teststatus", false)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(participants) != 0 {
			t.Errorf("unexpected number of participants found: %d, %d (have, want)", len(participants), 0)
		}
	})

}

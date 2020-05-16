package studydb

import (
	"testing"
	"time"

	"github.com/influenzanet/study-service/pkg/types"
)

func TestDbParticipantStateTest(t *testing.T) {
	testStudyKey := "teststudy234234"

	testPState := types.ParticipantState{
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
		_, err := testDBService.FindParticipantState(testInstanceID, testStudyKey, testPState.ParticipantID)
		if err == nil {
			t.Error("should return an error")
		}
	})

	t.Run("Testing update participant state, when not existing", func(t *testing.T) {
		pState, err := testDBService.SaveParticipantState(testInstanceID, testStudyKey, testPState)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if pState.ParticipantID != testPState.ParticipantID {
			t.Errorf("unexpected participantID -> have: %s, want: %s ", pState.ParticipantID, testPState.ParticipantID)
		}
	})

	t.Run("Testing find participant state, when existing", func(t *testing.T) {
		pState, err := testDBService.FindParticipantState(testInstanceID, testStudyKey, testPState.ParticipantID)
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
		pState, err := testDBService.SaveParticipantState(testInstanceID, testStudyKey, testPState)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if pState.StudyStatus != testPState.StudyStatus {
			t.Errorf("unexpected participantID -> have: %s, want: %s ", pState.StudyStatus, testPState.StudyStatus)
		}

		pState, err = testDBService.FindParticipantState(testInstanceID, testStudyKey, testPState.ParticipantID)
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

	pStates := []types.ParticipantState{
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
		_, err := testDBService.SaveParticipantState(testInstanceID, testStudyKey, ps)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
	}

	t.Run("Finding active status", func(t *testing.T) {
		participants, err := testDBService.FindParticipantsByStudyStatus(testInstanceID, testStudyKey, "active", false)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(participants) != 2 {
			t.Errorf("unexpected number of participants found: %d, %d (have, want)", len(participants), 2)
		}
	})

	t.Run("Finding inactive status ", func(t *testing.T) {
		participants, err := testDBService.FindParticipantsByStudyStatus(testInstanceID, testStudyKey, "inactive", true)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(participants) != 2 {
			t.Errorf("unexpected number of participants found: %d, %d (have, want)", len(participants), 2)
		}
	})

	t.Run("Findign not existing status", func(t *testing.T) {
		participants, err := testDBService.FindParticipantsByStudyStatus(testInstanceID, testStudyKey, "teststatus", false)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(participants) != 0 {
			t.Errorf("unexpected number of participants found: %d, %d (have, want)", len(participants), 0)
		}
	})

}

func TestFindAndExecuteOnParticipantsStates(t *testing.T) {
	testStudyKey := "teststudy_findandexecute"

	pStates := []types.ParticipantState{
		{
			ParticipantID: "1",
			StudyStatus:   "active",
			Flags: map[string]string{
				"test1": "1",
			},
		},
		{
			ParticipantID: "2",
			StudyStatus:   "active",
		},
	}

	for _, ps := range pStates {
		_, err := testDBService.SaveParticipantState(testInstanceID, testStudyKey, ps)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
	}

	t.Run("Finding inactive status ", func(t *testing.T) {
		err := testDBService.FindAndExecuteOnParticipantsStates(
			testInstanceID,
			testStudyKey,
			func(dbService *StudyDBService, p types.ParticipantState, instanceID, studyKey string) error {
				_, ok := p.Flags["test1"]
				if !ok {
					p.Flags = map[string]string{
						"test1": "1",
					}
				} else {
					p.Flags["test1"] = "newvalue"
				}
				_, err := dbService.SaveParticipantState(instanceID, studyKey, p)
				return err
			})
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}

		p, err := testDBService.FindParticipantState(testInstanceID, testStudyKey, pStates[0].ParticipantID)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		testval, ok := p.Flags["test1"]
		if !ok || testval != "newvalue" {
			t.Errorf("unexpected flags for p1: %s", p.Flags)
		}

		p, err = testDBService.FindParticipantState(testInstanceID, testStudyKey, pStates[1].ParticipantID)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		testval, ok = p.Flags["test1"]
		if !ok || testval != "1" {
			t.Errorf("unexpected flags for p2: %s", p.Flags)
		}
	})
}

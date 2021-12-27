package studydb

import (
	"context"
	"testing"
	"time"

	"github.com/influenzanet/study-service/pkg/types"
)

func TestDbParticipantStateTest(t *testing.T) {
	testStudyKey := "teststudy234234"

	testPState := types.ParticipantState{
		ParticipantID: "testPID0990",
		StudyStatus:   types.PARTICIPANT_STUDY_STATUS_ACTIVE,
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
			StudyStatus:   types.PARTICIPANT_STUDY_STATUS_ACTIVE,
		},
		{
			ParticipantID: "2",
			StudyStatus:   types.PARTICIPANT_STUDY_STATUS_ACTIVE,
		},
		{
			ParticipantID: "3",
			StudyStatus:   types.PARTICIPANT_STUDY_STATUS_EXITED,
		},
		{
			ParticipantID: "4",
			StudyStatus:   types.PARTICIPANT_STUDY_STATUS_EXITED,
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
		participants, err := testDBService.FindParticipantsByStudyStatus(testInstanceID, testStudyKey, types.PARTICIPANT_STUDY_STATUS_ACTIVE, false)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(participants) != 2 {
			t.Errorf("unexpected number of participants found: %d, %d (have, want)", len(participants), 2)
		}
	})

	t.Run("Finding exited status ", func(t *testing.T) {
		participants, err := testDBService.FindParticipantsByStudyStatus(testInstanceID, testStudyKey, types.PARTICIPANT_STUDY_STATUS_EXITED, true)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(participants) != 2 {
			t.Errorf("unexpected number of participants found: %d, %d (have, want)", len(participants), 2)
		}
	})

	t.Run("Finding not existing status", func(t *testing.T) {
		participants, err := testDBService.FindParticipantsByStudyStatus(testInstanceID, testStudyKey, "teststatus", false)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(participants) != 0 {
			t.Errorf("unexpected number of participants found: %d, %d (have, want)", len(participants), 0)
		}
	})

	t.Run("Finding not existing status", func(t *testing.T) {
		count, err := testDBService.GetParticipantCountByStatus(testInstanceID, testStudyKey, types.PARTICIPANT_STUDY_STATUS_ACTIVE)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if count != 2 {
			t.Errorf("unexpected number of active participants: %d", count)
			return
		}
	})
}

func TestFindAndExecuteOnParticipantsStates(t *testing.T) {
	testStudyKey := "teststudy_findandexecute"

	pStates := []types.ParticipantState{
		{
			ParticipantID: "1",
			StudyStatus:   types.PARTICIPANT_STUDY_STATUS_ACTIVE,
			Flags: map[string]string{
				"test1": "1",
			},
		},
		{
			ParticipantID: "2",
			StudyStatus:   types.PARTICIPANT_STUDY_STATUS_ACTIVE,
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
		ctx := context.Background()
		err := testDBService.FindAndExecuteOnParticipantsStates(
			ctx,
			testInstanceID,
			testStudyKey,
			types.STUDY_STATUS_ACTIVE,
			func(dbService *StudyDBService, p types.ParticipantState, instanceID, studyKey string, args ...interface{}) error {
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
			}, nil)
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

func TestDeleteMessagesFromParticipant(t *testing.T) {
	testStudyKey := "teststudy_deletemessages"

	pStates := []types.ParticipantState{
		{
			ParticipantID: "1",
		},
		{
			ParticipantID: "2",
			Messages: []types.ParticipantMessage{
				{
					ID:   "m1",
					Type: "test1",
				},
				{
					ID:   "m2",
					Type: "test1",
				},
				{
					ID:   "m3",
					Type: "test2",
				},
			},
		},
	}

	for _, ps := range pStates {
		_, err := testDBService.SaveParticipantState(testInstanceID, testStudyKey, ps)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
	}

	t.Run("for participant without messages ", func(t *testing.T) {
		pid := pStates[0].ParticipantID
		err := testDBService.DeleteMessagesFromParticipant(testInstanceID, testStudyKey, pid, []string{"m1", "m2"})
		if err == nil {
			t.Error("should return error")
			return
		}

		pState, err := testDBService.FindParticipantState(testInstanceID, testStudyKey, pid)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(pState.Messages) != 0 {
			t.Errorf("unexpected pState: %v", pState)
		}
	})

	t.Run("for participant with messages ", func(t *testing.T) {
		pid := pStates[1].ParticipantID
		err := testDBService.DeleteMessagesFromParticipant(testInstanceID, testStudyKey, pid, []string{"m1", "m2"})
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}

		pState, err := testDBService.FindParticipantState(testInstanceID, testStudyKey, pid)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(pState.Messages) != 1 {
			t.Errorf("unexpected pState: %v", pState)
		}
	})
}

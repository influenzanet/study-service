package studydb

import (
	"log"
	"testing"
	"time"

	"github.com/influenzanet/study-service/pkg/types"
)

func TestDbAddReport(t *testing.T) {
	testStudy := "testosaveresponse"
	testResp := types.Report{
		Key:           "test",
		ParticipantID: "test",
		Timestamp:     time.Now().Unix(),
	}
	t.Run("saving response", func(t *testing.T) {
		err := testDBService.SaveReport(testInstanceID, testStudy, testResp)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
	})
}

func TestDbFindReportForParticipant(t *testing.T) {
	testStudyKey := "teststudy_for_finding_responses"

	surveyResps := []types.Report{
		// mix participants and order for submittedAt
		{Key: "s1", ParticipantID: "u1", Timestamp: time.Now().Add(-30 * time.Hour * 24).Unix()},
		{Key: "s2", ParticipantID: "u1", Timestamp: time.Now().Add(-32 * time.Hour * 24).Unix()},
		{Key: "s1", ParticipantID: "u2", Timestamp: time.Now().Add(-29 * time.Hour * 24).Unix()},
		{Key: "s1", ParticipantID: "u1", Timestamp: time.Now().Add(-23 * time.Hour * 24).Unix()},
		{Key: "s1", ParticipantID: "u1", Timestamp: time.Now().Add(-6 * time.Hour * 24).Unix()},
		{Key: "s2", ParticipantID: "u1", Timestamp: time.Now().Add(-5 * time.Hour * 24).Unix()},
		{Key: "s1", ParticipantID: "u2", Timestamp: time.Now().Add(-6 * time.Hour * 24).Unix()},
		{Key: "s2", ParticipantID: "u2", Timestamp: time.Now().Add(-7 * time.Hour * 24).Unix()},
		{Key: "s1", ParticipantID: "u1", Timestamp: time.Now().Add(-15 * time.Hour * 24).Unix()},
		{Key: "s2", ParticipantID: "u1", Timestamp: time.Now().Add(-14 * time.Hour * 24).Unix()},
	}
	for _, sr := range surveyResps {
		err := testDBService.SaveReport(testInstanceID, testStudyKey, sr)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
	}

	t.Run("not existing participant", func(t *testing.T) {
		q := ReportQuery{
			ParticipantID: "u3",
		}
		responses, err := testDBService.FindReports(testInstanceID, testStudyKey, q)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(responses) > 0 {
			t.Errorf("unexpected number of responses found: %d", len(responses))
		}
	})

	t.Run("find last 2 report with key", func(t *testing.T) {
		q := ReportQuery{
			ParticipantID: "u1",
			Key:           "s1",
			Limit:         2,
		}
		responses, err := testDBService.FindReports(testInstanceID, testStudyKey, q)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(responses) != 2 {
			t.Errorf("unexpected number of responses found: %d", len(responses))
			return
		}
		if responses[0].Key != responses[1].Key && responses[0].Key != q.Key {
			t.Errorf("unexpected survey key: %s", responses[0].Key)
		}
	})

	t.Run("find last 2 report without key", func(t *testing.T) {
		q := ReportQuery{
			ParticipantID: "u1",
			Limit:         2,
		}
		responses, err := testDBService.FindReports(testInstanceID, testStudyKey, q)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(responses) != 2 {
			t.Errorf("unexpected number of responses found: %d", len(responses))
			return
		}
		if responses[0].Key == responses[1].Key {
			t.Errorf("unexpected survey keys: %s, %s", responses[0].Key, responses[1].Key)
		}
	})

	t.Run("find report after timestamp with key", func(t *testing.T) {
		q := ReportQuery{
			ParticipantID: "u1",
			Key:           "s2",
			Since:         time.Now().Add(-20 * time.Hour * 24).Unix(),
		}
		responses, err := testDBService.FindReports(testInstanceID, testStudyKey, q)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(responses) != 2 {
			t.Errorf("unexpected number of responses found: %d", len(responses))
			for _, r := range responses {
				log.Println(r)
			}
			return
		}
	})

	t.Run("find report after timestamp without key", func(t *testing.T) {
		q := ReportQuery{
			ParticipantID: "u1",
			Since:         time.Now().Add(-20 * time.Hour * 24).Unix(),
		}
		responses, err := testDBService.FindReports(testInstanceID, testStudyKey, q)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(responses) != 4 {
			t.Errorf("unexpected number of responses found: %d", len(responses))
			return
		}
	})
}

func TestDbUpdateParticipantIDonReports(t *testing.T) {
	testStudyKey := "teststudy_for_updating_pid"

	surveyResps := []types.Report{
		// mix participants and order for submittedAt
		{Key: "s1", ParticipantID: "u1", Timestamp: time.Now().Add(-30 * time.Hour * 24).Unix()},
		{Key: "s2", ParticipantID: "u1", Timestamp: time.Now().Add(-32 * time.Hour * 24).Unix()},
		{Key: "s1", ParticipantID: "u2", Timestamp: time.Now().Add(-29 * time.Hour * 24).Unix()},
		{Key: "s1", ParticipantID: "u2", Timestamp: time.Now().Add(-23 * time.Hour * 24).Unix()},
	}
	for _, sr := range surveyResps {
		err := testDBService.SaveReport(testInstanceID, testStudyKey, sr)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
	}

	t.Run("not existing participant", func(t *testing.T) {
		count, err := testDBService.UpdateParticipantIDonReports(testInstanceID, testStudyKey, "u3", "n3")
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if count > 0 {
			t.Errorf("unexpected number of modified respones found: %d", count)
		}
	})

	t.Run("u2", func(t *testing.T) {
		count, err := testDBService.UpdateParticipantIDonReports(testInstanceID, testStudyKey, "u2", "u3")
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if count != 2 {
			t.Errorf("unexpected number of modified respones found: %d", count)
		}

		r, _ := testDBService.FindReports(testInstanceID, testStudyKey, ReportQuery{
			ParticipantID: "u3",
		})
		if len(r) != 2 {
			t.Errorf("unexpected number of respones found: %d", len(r))
		}
	})
}

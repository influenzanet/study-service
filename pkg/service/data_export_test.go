package service

import (
	"context"
	"testing"

	"github.com/influenzanet/study-service/pkg/api"
	"github.com/influenzanet/study-service/pkg/dbs/studydb"
	"github.com/influenzanet/study-service/pkg/types"
)

func addTestSurveyResponses(studyDBservice *studydb.StudyDBService, instID string, studyKey string, repsonses []types.SurveyResponse) error {
	for _, resp := range repsonses {
		err := studyDBservice.AddSurveyResponse(testInstanceID, studyKey, resp)
		if err != nil {
			return err
		}
	}
	return nil
}

func TestGetStudyResponseStatisticsEndpoint(t *testing.T) {
	s := studyServiceServer{
		globalDBService:   testGlobalDBService,
		studyDBservice:    testStudyDBService,
		StudyGlobalSecret: "globsecretfortest1234",
	}

	testStudyKey := "testStudyfor_getsurveyresponsestatistics"
	testUser := "testuser"
	testStudy := types.Study{
		Key: testStudyKey,
		Members: []types.StudyMember{
			{
				UserID: testUser,
				Role:   "maintainer",
			},
		},
	}

	_, err := testStudyDBService.CreateStudy(testInstanceID, testStudy)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}

	// study for p1 key1 10
	err = addTestSurveyResponses(s.studyDBservice, testInstanceID, testStudyKey, []types.SurveyResponse{
		{Key: "key1", ParticipantID: "p1", SubmittedAt: 10},
		{Key: "key2", ParticipantID: "p1", SubmittedAt: 15},
		{Key: "key1", ParticipantID: "p1", SubmittedAt: 20},
		{Key: "key1", ParticipantID: "p2", SubmittedAt: 8},
		{Key: "key1", ParticipantID: "p2", SubmittedAt: 12},
		{Key: "key2", ParticipantID: "p2", SubmittedAt: 22},
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	t.Run("with missing request", func(t *testing.T) {
		_, err := s.GetStudyResponseStatistics(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty request", func(t *testing.T) {
		_, err := s.GetStudyResponseStatistics(context.Background(), &api.SurveyResponseQuery{})
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("as non study member", func(t *testing.T) {
		_, err := s.GetStudyResponseStatistics(context.Background(), &api.SurveyResponseQuery{
			Token: &api.TokenInfos{
				Id:         testUser + "wrong",
				InstanceId: testInstanceID,
			},
			StudyKey: testStudyKey,
		})
		ok, msg := shouldHaveGrpcErrorStatus(err, "not authorized to access this study")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("without query", func(t *testing.T) {
		resp, err := s.GetStudyResponseStatistics(context.Background(), &api.SurveyResponseQuery{
			Token: &api.TokenInfos{
				Id:         testUser,
				InstanceId: testInstanceID,
			},
			StudyKey: testStudyKey,
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		k1c, ok := resp.SurveyResponseCounts["key1"]
		if !ok || k1c != 4 {
			t.Errorf("unexpected number for k1: %d", k1c)
			return
		}
		k2c, ok := resp.SurveyResponseCounts["key2"]
		if !ok || k2c != 2 {
			t.Errorf("unexpected number for k2: %d", k1c)
			return
		}
	})

	t.Run("with from", func(t *testing.T) {
		resp, err := s.GetStudyResponseStatistics(context.Background(), &api.SurveyResponseQuery{
			Token: &api.TokenInfos{
				Id:         testUser,
				InstanceId: testInstanceID,
			},
			StudyKey: testStudyKey,
			From:     14,
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		k1c, ok := resp.SurveyResponseCounts["key1"]
		if !ok || k1c != 1 {
			t.Errorf("unexpected number for k1: %d", k1c)
			return
		}
		k2c, ok := resp.SurveyResponseCounts["key2"]
		if !ok || k2c != 2 {
			t.Errorf("unexpected number for k2: %d", k2c)
			return
		}
	})

	t.Run("with until", func(t *testing.T) {
		resp, err := s.GetStudyResponseStatistics(context.Background(), &api.SurveyResponseQuery{
			Token: &api.TokenInfos{
				Id:         testUser,
				InstanceId: testInstanceID,
			},
			StudyKey: testStudyKey,
			Until:    14,
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		k1c, ok := resp.SurveyResponseCounts["key1"]
		if !ok || k1c != 3 {
			t.Errorf("unexpected number for k1: %d", k1c)
			return
		}
		k2c, ok := resp.SurveyResponseCounts["key2"]
		if ok || k2c != 0 {
			t.Errorf("unexpected number for k2: %d", k2c)
			return
		}
	})

	t.Run("with time range", func(t *testing.T) {
		resp, err := s.GetStudyResponseStatistics(context.Background(), &api.SurveyResponseQuery{
			Token: &api.TokenInfos{
				Id:         testUser,
				InstanceId: testInstanceID,
			},
			StudyKey: testStudyKey,
			From:     11,
			Until:    19,
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		k1c, ok := resp.SurveyResponseCounts["key1"]
		if !ok || k1c != 1 {
			t.Errorf("unexpected number for k1: %d", k1c)
			return
		}
		k2c, ok := resp.SurveyResponseCounts["key2"]
		if !ok || k2c != 1 {
			t.Errorf("unexpected number for k2: %d", k2c)
			return
		}
	})
}

func TestStreamStudyResponsesEndpoint(t *testing.T) {
	s := studyServiceServer{
		globalDBService:   testGlobalDBService,
		studyDBservice:    testStudyDBService,
		StudyGlobalSecret: "globsecretfortest1234",
	}

	testStudyKey := "testStudyfor_streamsurveyresponses"
	testUser := "testuser"
	testStudy := types.Study{
		Key: testStudyKey,
		Members: []types.StudyMember{
			{
				UserID: testUser,
				Role:   "maintainer",
			},
		},
	}

	_, err := testStudyDBService.CreateStudy(testInstanceID, testStudy)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}

	// study for p1 key1 10
	err = addTestSurveyResponses(s.studyDBservice, testInstanceID, testStudyKey, []types.SurveyResponse{
		{Key: "key1", ParticipantID: "p1", SubmittedAt: 10},
		{Key: "key2", ParticipantID: "p1", SubmittedAt: 15},
		{Key: "key1", ParticipantID: "p1", SubmittedAt: 20},
		{Key: "key1", ParticipantID: "p2", SubmittedAt: 8},
		{Key: "key1", ParticipantID: "p2", SubmittedAt: 12},
		{Key: "key2", ParticipantID: "p2", SubmittedAt: 22},
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	// with nil
	// with empty
	// with non-study member
	// with spec. survey key
	// with from
	// with until
	// with all query params
	t.Error("test unimplemented")
}

package service

import (
	"context"
	"testing"

	"github.com/influenzanet/study-service/pkg/api"
	"github.com/influenzanet/study-service/pkg/types"
)

func TestCreateNewStudyEndpoint(t *testing.T) {
	s := studyServiceServer{
		globalDBService:   testGlobalDBService,
		studyDBservice:    testStudyDBService,
		StudyGlobalSecret: "globsecretfortest1234",
	}

	t.Run("with missing request", func(t *testing.T) {
		_, err := s.CreateNewStudy(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty request", func(t *testing.T) {
		_, err := s.CreateNewStudy(context.Background(), &api.NewStudyRequest{})
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	testStudy := api.Study{
		Key: "studyfor_creating",
	}

	t.Run("with missing user roles", func(t *testing.T) {
		_, err := s.CreateNewStudy(context.Background(), &api.NewStudyRequest{
			Token: &api.TokenInfos{
				InstanceId: testInstanceID,
				Id:         "user-id",
				Payload: map[string]string{
					"roles":    "PARTICIPANT",
					"username": "testuser",
				},
			},
			Study: &testStudy,
		})
		if err == nil {
			t.Error("should fail when don't have appropriate roles")
		}
	})

	t.Run("with correct user roles", func(t *testing.T) {
		study, err := s.CreateNewStudy(context.Background(), &api.NewStudyRequest{
			Token: &api.TokenInfos{
				InstanceId: testInstanceID,
				Id:         "user-id",
				Payload: map[string]string{
					"roles":    "PARTICIPANT,RESEARCHER",
					"username": "testuser",
				},
			},
			Study: &testStudy,
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(study.Members) != 1 {
			t.Errorf("unexpected number of members: %d", len(study.Members))
			return
		}
		if study.Members[0].Username != "testuser" {
			t.Errorf("unexpected member: %s", study.Members[0].Username)
		}
	})

	t.Run("with existing study key", func(t *testing.T) {
		_, err := s.CreateNewStudy(context.Background(), &api.NewStudyRequest{
			Token: &api.TokenInfos{
				InstanceId: testInstanceID,
				Id:         "user-id",
				Payload: map[string]string{
					"roles":    "PARTICIPANT,RESEARCHER",
					"username": "testuser",
				},
			},
			Study: &testStudy,
		})
		if err == nil {
			t.Error("should fail when study key already used")
		}
	})

}

func TestSaveSurveyToStudyEndpoint(t *testing.T) {
	s := studyServiceServer{
		globalDBService:   testGlobalDBService,
		studyDBservice:    testStudyDBService,
		StudyGlobalSecret: "globsecretfortest1234",
	}

	testUser := "testuser"
	testStudy := types.Study{
		Key: "testStudy_for_save_survey",
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

	t.Run("with missing request", func(t *testing.T) {
		_, err := s.SaveSurveyToStudy(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty request", func(t *testing.T) {
		_, err := s.SaveSurveyToStudy(context.Background(), &api.AddSurveyReq{})
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with wrong user role", func(t *testing.T) {
		testSurvey := api.Survey{
			Current: &api.SurveyVersion{
				SurveyDefinition: &api.SurveyItem{
					Key: "testkey",
				},
			},
		}
		_, err := s.SaveSurveyToStudy(context.Background(), &api.AddSurveyReq{
			Token: &api.TokenInfos{
				Id:         "user",
				InstanceId: testInstanceID,
				Payload: map[string]string{
					"roles":    "PARTICIPANT,RESEARCHER",
					"username": "testuserwrong",
				},
			},
			StudyKey: testStudy.Key,
			Survey:   &testSurvey,
		})
		if err == nil {
			t.Error("should fail")
		}
	})

	t.Run("with new survey to add", func(t *testing.T) {
		testSurvey := api.Survey{
			Current: &api.SurveyVersion{
				SurveyDefinition: &api.SurveyItem{
					Key: "testkey",
				},
			},
		}
		resp, err := s.SaveSurveyToStudy(context.Background(), &api.AddSurveyReq{
			Token: &api.TokenInfos{
				Id:         "testuser",
				InstanceId: testInstanceID,
				Payload: map[string]string{
					"roles":    "PARTICIPANT,RESEARCHER",
					"username": "testusername",
				},
			},
			StudyKey: testStudy.Key,
			Survey:   &testSurvey,
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if resp.Current.SurveyDefinition.Key != "testkey" {
			t.Error("unexpected survey key")
		}
	})
}

func TestRemoveSurveyFromStudyEndpoint(t *testing.T) {
	s := studyServiceServer{
		globalDBService:   testGlobalDBService,
		studyDBservice:    testStudyDBService,
		StudyGlobalSecret: "globsecretfortest1234",
	}

	testStudyKey := "testStudyfor_removesurveys"
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

	testSurveys := []types.Survey{
		{Current: types.SurveyVersion{SurveyDefinition: types.SurveyItem{Key: "1"}}},
		{Current: types.SurveyVersion{SurveyDefinition: types.SurveyItem{Key: "3"}}},
		{Current: types.SurveyVersion{SurveyDefinition: types.SurveyItem{Key: "2"}}},
	}
	for _, s := range testSurveys {
		_, err := testStudyDBService.SaveSurvey(testInstanceID, testStudyKey, s)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
	}

	t.Run("with missing request", func(t *testing.T) {
		_, err := s.RemoveSurveyFromStudy(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty request", func(t *testing.T) {
		_, err := s.RemoveSurveyFromStudy(context.Background(), &api.SurveyReferenceRequest{})
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})
	t.Run("as non study member", func(t *testing.T) {
		_, err := s.RemoveSurveyFromStudy(context.Background(), &api.SurveyReferenceRequest{
			Token: &api.TokenInfos{
				Id:         testUser + "wrong",
				InstanceId: testInstanceID,
			},
			StudyKey:  testStudyKey,
			SurveyKey: "1",
		})
		ok, msg := shouldHaveGrpcErrorStatus(err, "")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with not existing survey", func(t *testing.T) {
		_, err := s.RemoveSurveyFromStudy(context.Background(), &api.SurveyReferenceRequest{
			Token: &api.TokenInfos{
				Id:         testUser,
				InstanceId: testInstanceID,
			},
			StudyKey:  testStudyKey + "wrong",
			SurveyKey: "1",
		})
		ok, msg := shouldHaveGrpcErrorStatus(err, "")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with correct inputs", func(t *testing.T) {
		_, err := s.RemoveSurveyFromStudy(context.Background(), &api.SurveyReferenceRequest{
			Token: &api.TokenInfos{
				Id:         testUser,
				InstanceId: testInstanceID,
			},
			StudyKey:  testStudyKey,
			SurveyKey: "1",
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		surveys, err := testStudyDBService.FindAllSurveyDefsForStudy(testInstanceID, testStudyKey, true)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(surveys) != 2 {
			t.Errorf("unexpected number of survey: %d", len(surveys))
		}
	})

}

func TestGetStudySurveyInfosEndpoint(t *testing.T) {
	s := studyServiceServer{
		globalDBService:   testGlobalDBService,
		studyDBservice:    testStudyDBService,
		StudyGlobalSecret: "globsecretfortest1234",
	}

	testStudyKey := "testStudyfor_finding_all_surveys"

	testSurveys := []types.Survey{
		{Current: types.SurveyVersion{SurveyDefinition: types.SurveyItem{Key: "1"}}},
		{Current: types.SurveyVersion{SurveyDefinition: types.SurveyItem{Key: "3"}}},
		{Current: types.SurveyVersion{SurveyDefinition: types.SurveyItem{Key: "2"}}},
	}
	for _, s := range testSurveys {
		_, err := testStudyDBService.SaveSurvey(testInstanceID, testStudyKey, s)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
	}

	t.Run("with missing request", func(t *testing.T) {
		_, err := s.GetStudySurveyInfos(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty request", func(t *testing.T) {
		_, err := s.GetStudySurveyInfos(context.Background(), &api.StudyReferenceReq{})
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("correct args", func(t *testing.T) {
		surveys, err := s.GetStudySurveyInfos(context.Background(), &api.StudyReferenceReq{
			Token:    &api.TokenInfos{Id: "test", InstanceId: testInstanceID, ProfilId: "test"},
			StudyKey: testStudyKey,
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(surveys.Infos) != 3 {
			t.Errorf("unexpected number of surveys: %d", len(surveys.Infos))
		}
	})
}

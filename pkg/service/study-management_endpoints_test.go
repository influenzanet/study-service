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

func TestGetAllStudiesEndpoint(t *testing.T) {
	s := studyServiceServer{
		globalDBService:   testGlobalDBService,
		studyDBservice:    testStudyDBService,
		StudyGlobalSecret: "globsecretfortest1234",
	}

	testStudyKey := "testStudyfor_getallstudies"
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

	t.Run("with missing request", func(t *testing.T) {
		_, err := s.GetAllStudies(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty request", func(t *testing.T) {
		_, err := s.GetAllStudies(context.Background(), &api.TokenInfos{})
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with non admin user", func(t *testing.T) {
		_, err := s.GetAllStudies(context.Background(), &api.TokenInfos{
			Id:         "user",
			InstanceId: testInstanceID,
			Payload: map[string]string{
				"roles":    "PARTICIPANT",
				"username": "testuser",
			},
		})
		ok, msg := shouldHaveGrpcErrorStatus(err, "not authorized")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with researcher user", func(t *testing.T) {
		resp, err := s.GetAllStudies(context.Background(), &api.TokenInfos{
			Id:         "user",
			InstanceId: testInstanceID,
			Payload: map[string]string{
				"roles":    "PARTICIPANT,RESEARCHER",
				"username": "testuser",
			},
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(resp.Studies) < 1 {
			t.Error("at least one study should be there")
		}
	})
}

func TestGetStudyEndpoint(t *testing.T) {
	s := studyServiceServer{
		globalDBService:   testGlobalDBService,
		studyDBservice:    testStudyDBService,
		StudyGlobalSecret: "globsecretfortest1234",
	}

	testStudyKey := "testStudyfor_getstudy"
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

	t.Run("with missing request", func(t *testing.T) {
		_, err := s.GetStudy(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty request", func(t *testing.T) {
		_, err := s.GetStudy(context.Background(), &api.StudyReferenceReq{})
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with non admin user", func(t *testing.T) {
		_, err := s.GetStudy(context.Background(), &api.StudyReferenceReq{
			Token: &api.TokenInfos{
				Id:         "user",
				InstanceId: testInstanceID,
				Payload: map[string]string{
					"roles":    "PARTICIPANT",
					"username": "testuser",
				},
			},
			StudyKey: testStudyKey,
		})
		ok, msg := shouldHaveGrpcErrorStatus(err, "not authorized")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with researcher user", func(t *testing.T) {
		resp, err := s.GetStudy(context.Background(), &api.StudyReferenceReq{
			Token: &api.TokenInfos{
				Id:         "user",
				InstanceId: testInstanceID,
				Payload: map[string]string{
					"roles":    "PARTICIPANT,RESEARCHER",
					"username": "testuser",
				},
			},
			StudyKey: testStudyKey,
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(resp.Members) < 1 {
			t.Error("at least one study member should be there")
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

func TestGetSurveyDefForStudyEndpoint(t *testing.T) {
	s := studyServiceServer{
		globalDBService:   testGlobalDBService,
		studyDBservice:    testStudyDBService,
		StudyGlobalSecret: "globsecretfortest1234",
	}

	testStudyKey := "testStudyfor_getsurveydef"
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
		_, err := s.GetSurveyDefForStudy(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty request", func(t *testing.T) {
		_, err := s.GetSurveyDefForStudy(context.Background(), &api.SurveyReferenceRequest{})
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})
	t.Run("as non study member", func(t *testing.T) {
		_, err := s.GetSurveyDefForStudy(context.Background(), &api.SurveyReferenceRequest{
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
		_, err := s.GetSurveyDefForStudy(context.Background(), &api.SurveyReferenceRequest{
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
		resp, err := s.GetSurveyDefForStudy(context.Background(), &api.SurveyReferenceRequest{
			Token: &api.TokenInfos{
				Id:         testUser,
				InstanceId: testInstanceID,
			},
			StudyKey:  testStudyKey,
			SurveyKey: "1",
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if resp.Current.SurveyDefinition.Key != "1" {
			t.Errorf("unexpected survey def: %v", resp)
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

func TestSaveStudyMemberEndpoint(t *testing.T) {
	s := studyServiceServer{
		globalDBService:   testGlobalDBService,
		studyDBservice:    testStudyDBService,
		StudyGlobalSecret: "globsecretfortest1234",
	}

	testStudyKey := "testStudyfor_savemember"
	testUserID := "testuserid"
	testStudy := types.Study{
		Key: testStudyKey,
		Members: []types.StudyMember{
			{
				UserID: testUserID,
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
		_, err := s.SaveStudyMember(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty request", func(t *testing.T) {
		_, err := s.SaveStudyMember(context.Background(), &api.StudyMemberReq{})
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with non study member non admin user", func(t *testing.T) {
		_, err := s.SaveStudyMember(context.Background(), &api.StudyMemberReq{
			Token: &api.TokenInfos{
				Id:         "user",
				InstanceId: testInstanceID,
				Payload: map[string]string{
					"roles":    "PARTICIPANT",
					"username": "testuser2",
				},
			},
			StudyKey: testStudyKey,
			Member: &api.Study_Member{
				Role:     "analyst",
				UserId:   "newid",
				Username: "new user",
			},
		})
		ok, msg := shouldHaveGrpcErrorStatus(err, "not authorized to access this study")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with non study member but admin user", func(t *testing.T) {
		resp, err := s.SaveStudyMember(context.Background(), &api.StudyMemberReq{
			Token: &api.TokenInfos{
				Id:         "user",
				InstanceId: testInstanceID,
				Payload: map[string]string{
					"roles":    "PARTICIPANT,RESEARCHER,ADMIN",
					"username": "testuser2",
				},
			},
			StudyKey: testStudyKey,
			Member: &api.Study_Member{
				Role:     "maintainer",
				UserId:   "newid",
				Username: "new user",
			},
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(resp.Members) != 2 {
			t.Error("unexpected number of members")
		}
	})

	t.Run("with study member", func(t *testing.T) {
		resp, err := s.SaveStudyMember(context.Background(), &api.StudyMemberReq{
			Token: &api.TokenInfos{
				Id:         testUserID,
				InstanceId: testInstanceID,
				Payload: map[string]string{
					"roles":    "PARTICIPANT,RESEARCHER,ADMIN",
					"username": "testuser",
				},
			},
			StudyKey: testStudyKey,
			Member: &api.Study_Member{
				Role:     "test",
				UserId:   "newid",
				Username: "new user",
			},
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(resp.Members) != 2 {
			t.Errorf("unexpected number of members: %d", len(resp.Members))
			return
		}
		if resp.Members[1].Role != "test" {
			t.Error("unexpected role in updated member")
		}
	})
}

func TestRemoveStudyMemberEndpoint(t *testing.T) {
	s := studyServiceServer{
		globalDBService:   testGlobalDBService,
		studyDBservice:    testStudyDBService,
		StudyGlobalSecret: "globsecretfortest1234",
	}

	testStudyKey := "testStudyfor_removemember"
	testUserID := "testuserid"
	testStudy := types.Study{
		Key: testStudyKey,
		Members: []types.StudyMember{
			{
				UserID: testUserID,
				Role:   "maintainer",
			},
			{
				UserID: "userid2",
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
		_, err := s.RemoveStudyMember(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty request", func(t *testing.T) {
		_, err := s.RemoveStudyMember(context.Background(), &api.StudyMemberReq{})
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with non study member", func(t *testing.T) {
		_, err := s.RemoveStudyMember(context.Background(), &api.StudyMemberReq{
			Token: &api.TokenInfos{
				Id:         "user",
				InstanceId: testInstanceID,
				Payload: map[string]string{
					"roles":    "PARTICIPANT",
					"username": "testuser2",
				},
			},
			StudyKey: testStudyKey,
			Member: &api.Study_Member{
				UserId: "userid2",
			},
		})
		ok, msg := shouldHaveGrpcErrorStatus(err, "not authorized to access this study")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with study member", func(t *testing.T) {
		resp, err := s.SaveStudyMember(context.Background(), &api.StudyMemberReq{
			Token: &api.TokenInfos{
				Id:         testUserID,
				InstanceId: testInstanceID,
				Payload: map[string]string{
					"roles":    "PARTICIPANT,RESEARCHER,ADMIN",
					"username": "testuser",
				},
			},
			StudyKey: testStudyKey,
			Member: &api.Study_Member{
				UserId: "userid2",
			},
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(resp.Members) != 1 {
			t.Error("unexpected number of members")
		}
	})
}

func TestSaveStudyRulesEndpoint(t *testing.T) {
	s := studyServiceServer{
		globalDBService:   testGlobalDBService,
		studyDBservice:    testStudyDBService,
		StudyGlobalSecret: "globsecretfortest1234",
	}

	testStudyKey := "testStudyfor_saverules"
	testUserID := "testuserid"
	testStudy := types.Study{
		Key: testStudyKey,
		Members: []types.StudyMember{
			{
				UserID: testUserID,
				Role:   "maintainer",
			},
		},
		Rules: []types.Expression{},
	}

	_, err := testStudyDBService.CreateStudy(testInstanceID, testStudy)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}

	t.Run("with missing request", func(t *testing.T) {
		_, err := s.SaveStudyRules(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty request", func(t *testing.T) {
		_, err := s.SaveStudyRules(context.Background(), &api.StudyRulesReq{})
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with non study member", func(t *testing.T) {
		_, err := s.SaveStudyRules(context.Background(), &api.StudyRulesReq{
			Token: &api.TokenInfos{
				Id:         "user",
				InstanceId: testInstanceID,
				Payload: map[string]string{
					"roles":    "PARTICIPANT",
					"username": "testuser2",
				},
			},
			StudyKey: testStudyKey,
			Rules: []*api.Expression{
				{Name: "test"},
			},
		})
		ok, msg := shouldHaveGrpcErrorStatus(err, "not authorized to access this study")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with study member", func(t *testing.T) {
		resp, err := s.SaveStudyRules(context.Background(), &api.StudyRulesReq{
			Token: &api.TokenInfos{
				Id:         testUserID,
				InstanceId: testInstanceID,
				Payload: map[string]string{
					"roles":    "PARTICIPANT,RESEARCHER,ADMIN",
					"username": "testuser",
				},
			},
			StudyKey: testStudyKey,
			Rules: []*api.Expression{
				{Name: "test"},
			},
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(resp.Rules) != 1 {
			t.Error("unexpected number of rules")
		}
	})
}

func TestSaveStudyStatusEndpoint(t *testing.T) {
	s := studyServiceServer{
		globalDBService:   testGlobalDBService,
		studyDBservice:    testStudyDBService,
		StudyGlobalSecret: "globsecretfortest1234",
	}

	testStudyKey := "testStudyfor_savestatus"
	testUserID := "testuserid"
	testStudy := types.Study{
		Key: testStudyKey,
		Members: []types.StudyMember{
			{
				UserID: testUserID,
				Role:   "maintainer",
			},
		},
		Status: "active",
	}

	_, err := testStudyDBService.CreateStudy(testInstanceID, testStudy)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}

	t.Run("with missing request", func(t *testing.T) {
		_, err := s.SaveStudyStatus(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty request", func(t *testing.T) {
		_, err := s.SaveStudyStatus(context.Background(), &api.StudyStatusReq{})
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with non study member", func(t *testing.T) {
		_, err := s.SaveStudyStatus(context.Background(), &api.StudyStatusReq{
			Token: &api.TokenInfos{
				Id:         "user",
				InstanceId: testInstanceID,
				Payload: map[string]string{
					"roles":    "PARTICIPANT",
					"username": "testuser2",
				},
			},
			StudyKey:  testStudyKey,
			NewStatus: "test",
		})
		ok, msg := shouldHaveGrpcErrorStatus(err, "not authorized to access this study")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with study member", func(t *testing.T) {
		resp, err := s.SaveStudyStatus(context.Background(), &api.StudyStatusReq{
			Token: &api.TokenInfos{
				Id:         testUserID,
				InstanceId: testInstanceID,
				Payload: map[string]string{
					"roles":    "PARTICIPANT,RESEARCHER,ADMIN",
					"username": "testuser",
				},
			},
			StudyKey:  testStudyKey,
			NewStatus: "test",
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if resp.Status != "test" {
			t.Error("unexpected status")
		}
	})
}

func TestSaveStudyPropsEndpoint(t *testing.T) {
	s := studyServiceServer{
		globalDBService:   testGlobalDBService,
		studyDBservice:    testStudyDBService,
		StudyGlobalSecret: "globsecretfortest1234",
	}

	testStudyKey := "testStudyfor_saveprops"
	testUserID := "testuserid"
	testStudy := types.Study{
		Key: testStudyKey,
		Members: []types.StudyMember{
			{
				UserID: testUserID,
				Role:   "maintainer",
			},
		},
		Status: "active",
		Props: types.StudyProps{
			Name: []types.LocalisedObject{
				{Code: "en "},
			},
		},
	}

	_, err := testStudyDBService.CreateStudy(testInstanceID, testStudy)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}

	t.Run("with missing request", func(t *testing.T) {
		_, err := s.SaveStudyProps(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty request", func(t *testing.T) {
		_, err := s.SaveStudyProps(context.Background(), &api.StudyPropsReq{})
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with non study member", func(t *testing.T) {
		_, err := s.SaveStudyProps(context.Background(), &api.StudyPropsReq{
			Token: &api.TokenInfos{
				Id:         "user",
				InstanceId: testInstanceID,
				Payload: map[string]string{
					"roles":    "PARTICIPANT",
					"username": "testuser2",
				},
			},
			StudyKey: testStudyKey,
			Props: &api.Study_Props{
				Name: []*api.LocalisedObject{
					{Code: "en"},
					{Code: "de"},
				},
			},
		})
		ok, msg := shouldHaveGrpcErrorStatus(err, "not authorized to access this study")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with study member", func(t *testing.T) {
		resp, err := s.SaveStudyProps(context.Background(), &api.StudyPropsReq{
			Token: &api.TokenInfos{
				Id:         testUserID,
				InstanceId: testInstanceID,
				Payload: map[string]string{
					"roles":    "PARTICIPANT,RESEARCHER,ADMIN",
					"username": "testuser",
				},
			},
			StudyKey: testStudyKey,
			Props: &api.Study_Props{
				Name: []*api.LocalisedObject{
					{Code: "en"},
					{Code: "de"},
				},
			},
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(resp.Props.Name) != 2 {
			t.Error("unexpected name loc objs")
		}
	})
}

func TestDeleteStudyEndpoint(t *testing.T) {
	s := studyServiceServer{
		globalDBService:   testGlobalDBService,
		studyDBservice:    testStudyDBService,
		StudyGlobalSecret: "globsecretfortest1234",
	}

	testStudyKey := "testStudyfor_todelete"
	testUserID := "testuserid"
	testStudy := types.Study{
		Key: testStudyKey,
		Members: []types.StudyMember{
			{
				UserID: testUserID,
				Role:   "maintainer",
			},
		},
		Status: "active",
		Props: types.StudyProps{
			Name: []types.LocalisedObject{
				{Code: "en "},
			},
		},
	}

	_, err := testStudyDBService.CreateStudy(testInstanceID, testStudy)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}

	t.Run("with missing request", func(t *testing.T) {
		_, err := s.DeleteStudy(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty request", func(t *testing.T) {
		_, err := s.DeleteStudy(context.Background(), &api.StudyReferenceReq{})
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with non study member", func(t *testing.T) {
		_, err := s.DeleteStudy(context.Background(), &api.StudyReferenceReq{
			Token: &api.TokenInfos{
				Id:         "user",
				InstanceId: testInstanceID,
				Payload: map[string]string{
					"roles":    "PARTICIPANT",
					"username": "testuser2",
				},
			},
			StudyKey: testStudyKey,
		})
		ok, msg := shouldHaveGrpcErrorStatus(err, "not authorized to access this study")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with study member", func(t *testing.T) {
		_, err := s.DeleteStudy(context.Background(), &api.StudyReferenceReq{
			Token: &api.TokenInfos{
				Id:         testUserID,
				InstanceId: testInstanceID,
				Payload: map[string]string{
					"roles":    "PARTICIPANT,RESEARCHER,ADMIN",
					"username": "testuser",
				},
			},
			StudyKey: testStudyKey,
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		_, err = s.studyDBservice.GetStudyByStudyKey(testInstanceID, testStudyKey)
		if err == nil {
			t.Error("study should be deleted")
		}
	})
}

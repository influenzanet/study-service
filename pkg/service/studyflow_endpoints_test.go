package service

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/influenzanet/study-service/pkg/api"
	"github.com/influenzanet/study-service/pkg/types"
)

func TestCheckIfParticipantExists(t *testing.T) {
	// Test setup
	s := studyServiceServer{
		globalDBService:   testGlobalDBService,
		studyDBservice:    testStudyDBService,
		StudyGlobalSecret: "globsecretfortest1234",
	}
	testStudyKey := "teststudy_checkifparticipantexists"

	pStates := []types.ParticipantState{
		{
			ParticipantID: "1",
			StudyStatus:   "active",
		},
	}

	for _, ps := range pStates {
		_, err := testStudyDBService.SaveParticipantState(testInstanceID, testStudyKey, ps)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
	}

	// Tests
	t.Run("with existing participant", func(t *testing.T) {
		if !s.checkIfParticipantExists(testInstanceID, testStudyKey, "1") {
			t.Error("should be true if participant exists")
		}
	})

	t.Run("with not existing participant", func(t *testing.T) {
		if s.checkIfParticipantExists(testInstanceID, testStudyKey, "2") {
			t.Error("should be false if participant does not exist")
		}
	})
}

func TestGetAndPerformStudyRules(t *testing.T) {
	s := studyServiceServer{
		globalDBService:   testGlobalDBService,
		studyDBservice:    testStudyDBService,
		StudyGlobalSecret: "globsecretfortest1234",
	}

	testStudy := types.Study{
		Key:       "studytocheckifrulesareworking",
		SecretKey: "testsecret",
		Rules: []types.Expression{
			{
				Name: "IFTHEN",
				Data: []types.ExpressionArg{
					{
						DType: "exp",
						Exp: &types.Expression{
							Name: "checkEventType",
							Data: []types.ExpressionArg{
								{Str: "ENTER"},
							},
						},
					},
					{
						DType: "exp",
						Exp: &types.Expression{
							Name: "UPDATE_FLAG",
							Data: []types.ExpressionArg{
								{Str: "testKey"},
								{Str: "testValue"},
							},
						},
					},
				},
			},
			{
				Name: "IFTHEN",
				Data: []types.ExpressionArg{
					{
						DType: "exp",
						Exp: &types.Expression{
							Name: "checkEventType",
							Data: []types.ExpressionArg{
								{Str: "SUBMIT"},
							},
						},
					},
					{
						DType: "exp",
						Exp: &types.Expression{
							Name: "UPDATE_FLAG",
							Data: []types.ExpressionArg{
								{Str: "testKey"},
								{Str: "testValue2"},
							},
						},
					},
				},
			},
		},
	}

	testStudy, err := testStudyDBService.CreateStudy(testInstanceID, testStudy)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}

	pState := types.ParticipantState{
		ParticipantID: "1",
	}

	t.Run("ENTER event", func(t *testing.T) {
		testEvent := types.StudyEvent{
			Type: "ENTER",
		}

		pState, err = s.getAndPerformStudyRules(testInstanceID, testStudy.Key, pState, testEvent)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		v, ok := pState.Flags["testKey"]
		if !ok {
			t.Error("testKey not found")
		}
		if v != "testValue" {
			t.Errorf("testValue not matches %s", v)
		}
	})
	t.Run("SUBMIT event", func(t *testing.T) {
		testEvent := types.StudyEvent{
			Type: "SUBMIT",
			Response: types.SurveyResponse{
				Key: "testsurvey",
			},
		}

		pState, err = s.getAndPerformStudyRules(testInstanceID, testStudy.Key, pState, testEvent)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		v, ok := pState.Flags["testKey"]
		if !ok {
			t.Error("testKey not found")
		}
		if v != "testValue2" {
			t.Errorf("testValue not matches %s", v)
		}
	})
}

func TestEnterStudyEndpoint(t *testing.T) {
	s := studyServiceServer{
		globalDBService:   testGlobalDBService,
		studyDBservice:    testStudyDBService,
		StudyGlobalSecret: "globsecretfortest1234",
	}

	testStudy := types.Study{
		Key:       "studyfortestingenterstudy",
		SecretKey: "testsecret",
		Rules: []types.Expression{
			{
				Name: "IFTHEN",
				Data: []types.ExpressionArg{
					{
						DType: "exp",
						Exp: &types.Expression{
							Name: "checkEventType",
							Data: []types.ExpressionArg{
								{Str: "ENTER"},
							},
						},
					},
					{
						DType: "exp",
						Exp: &types.Expression{
							Name: "ADD_NEW_SURVEY",
							Data: []types.ExpressionArg{
								{Str: "testsurvey"},
								{DType: "num", Num: 0},
								{DType: "num", Num: 0},
							},
						},
					},
				},
			},
			{
				Name: "IFTHEN",
				Data: []types.ExpressionArg{
					{
						DType: "exp",
						Exp: &types.Expression{
							Name: "checkEventType",
							Data: []types.ExpressionArg{
								{Str: "SUBMIT"},
							},
						},
					},
					{
						DType: "exp",
						Exp: &types.Expression{
							Name: "UPDATE_FLAG",
							Data: []types.ExpressionArg{
								{Str: "testKey"},
								{Str: "testValue2"},
							},
						},
					},
				},
			},
		},
	}

	testStudy, err := testStudyDBService.CreateStudy(testInstanceID, testStudy)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}

	t.Run("with missing request", func(t *testing.T) {
		_, err := s.EnterStudy(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty request", func(t *testing.T) {
		_, err := s.EnterStudy(context.Background(), &api.EnterStudyRequest{})
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("wrong study key", func(t *testing.T) {
		req := &api.EnterStudyRequest{
			Token: &api.TokenInfos{
				Id:         "testuser",
				InstanceId: testInstanceID,
			},
			StudyKey: testStudy.Key + "wrong",
		}
		_, err := s.EnterStudy(context.Background(), req)
		if err == nil {
			t.Error("should return an error")
			return
		}
	})

	t.Run("correct values", func(t *testing.T) {
		req := &api.EnterStudyRequest{
			Token: &api.TokenInfos{
				Id:         "testuser",
				InstanceId: testInstanceID,
			},
			StudyKey: testStudy.Key,
		}
		resp, err := s.EnterStudy(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(resp.Surveys) != 1 {
			t.Errorf("unexpected number of surveys: %d", len(resp.Surveys))
		}
	})

	t.Run("existing participant (user) id", func(t *testing.T) {
		req := &api.EnterStudyRequest{
			Token: &api.TokenInfos{
				Id:         "testuser",
				InstanceId: testInstanceID,
			},
			StudyKey: testStudy.Key,
		}
		_, err := s.EnterStudy(context.Background(), req)
		if err == nil {
			t.Error("should return an error")
			return
		}
	})
}

func TestGetAssignedSurveysEndpoint(t *testing.T) {
	s := studyServiceServer{
		globalDBService:   testGlobalDBService,
		studyDBservice:    testStudyDBService,
		StudyGlobalSecret: "globsecretfortest1234",
	}

	studies := []types.Study{
		{
			Status:    "active",
			Key:       "studyforassignedsurvey1",
			SecretKey: "testsecret",
		},
		{
			Status:    "active",
			Key:       "studyforassignedsurvey2",
			SecretKey: "testsecret2",
		},
		{
			Status:    "active",
			Key:       "studyforassignedsurvey3",
			SecretKey: "testsecret3",
		},
	}

	for _, study := range studies {
		_, err := s.studyDBservice.CreateStudy(testInstanceID, study)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
	}

	testUserID := "234234laaabbb3423"

	pid1, err := s.userIDToParticipantID(testInstanceID, "studyforassignedsurvey1", testUserID)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	pid2, err := s.userIDToParticipantID(testInstanceID, "studyforassignedsurvey2", testUserID)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	pState1 := types.ParticipantState{
		ParticipantID: pid1,
		StudyStatus:   "active",
		AssignedSurveys: []types.AssignedSurvey{
			{SurveyKey: "s1"},
		},
	}
	pState2 := types.ParticipantState{
		ParticipantID: pid2,
		StudyStatus:   "active",
		AssignedSurveys: []types.AssignedSurvey{
			{SurveyKey: "s1"},
		},
	}

	_, err = testStudyDBService.SaveParticipantState(testInstanceID, "studyforassignedsurvey1", pState1)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	_, err = testStudyDBService.SaveParticipantState(testInstanceID, "studyforassignedsurvey2", pState2)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}

	t.Run("with missing request", func(t *testing.T) {
		_, err := s.GetAssignedSurveys(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty request", func(t *testing.T) {
		_, err := s.GetAssignedSurveys(context.Background(), &api.TokenInfos{})
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("wrong study key", func(t *testing.T) {
		resp, err := s.GetAssignedSurveys(context.Background(), &api.TokenInfos{
			Id:         testUserID,
			InstanceId: testInstanceID,
			ProfilId:   testUserID,
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(resp.Surveys) != 2 {
			t.Errorf("unexpected number of surveys: %d", len(resp.Surveys))
			return
		}
		if resp.Surveys[0].StudyKey != "studyforassignedsurvey1" || resp.Surveys[1].StudyKey != "studyforassignedsurvey2" {
			t.Error(resp)
		}
	})
}

func TestGetAssignedSurveyEndpoint(t *testing.T) {
	s := studyServiceServer{
		globalDBService:   testGlobalDBService,
		studyDBservice:    testStudyDBService,
		StudyGlobalSecret: "globsecretfortest1234",
	}

	testStudyKey := "teststudy_for_get_assignedsurvey"
	testUserID := "234234laaabbb3423"
	studies := []types.Study{
		{
			Status:    "active",
			Key:       testStudyKey,
			SecretKey: "testsecret",
		},
	}

	for _, study := range studies {
		_, err := testStudyDBService.CreateStudy(testInstanceID, study)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
	}

	pid1, err := s.userIDToParticipantID(testInstanceID, testStudyKey, testUserID)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	surveyResps := []types.SurveyResponse{
		// mix participants and order for submittedAt
		{Key: "s1", ParticipantID: pid1, SubmittedAt: time.Now().Add(-6 * time.Hour * 24).Unix(), Responses: []types.SurveyItemResponse{
			{Key: "s1.1"},
			{Key: "s1.2"},
			{Key: "s1.3"},
		}},
		{Key: "s2", ParticipantID: pid1, SubmittedAt: time.Now().Add(-5 * time.Hour * 24).Unix(), Responses: []types.SurveyItemResponse{
			{Key: "s2.1"},
			{Key: "s2.2"},
		}},
		{Key: "s1", ParticipantID: pid1, SubmittedAt: time.Now().Add(-15 * time.Hour * 24).Unix(), Responses: []types.SurveyItemResponse{
			{Key: "s1.1"},
			{Key: "s1.2"},
			{Key: "s1.3"},
		}},
		{Key: "s2", ParticipantID: pid1, SubmittedAt: time.Now().Add(-14 * time.Hour * 24).Unix(), Responses: []types.SurveyItemResponse{
			{Key: "s2.1"},
			{Key: "s2.2"},
			{Key: "s2.3"},
		}},
	}
	for _, sr := range surveyResps {
		err := testStudyDBService.AddSurveyResponse(testInstanceID, testStudyKey, sr)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
	}

	testSurvey := types.Survey{
		Current: types.SurveyVersion{
			SurveyDefinition: types.SurveyItem{
				Key: "t1",
			},
		},
		ContextRules: &types.SurveyContextDef{
			Mode: &types.ExpressionArg{Str: "test"},
			PreviousResponses: []types.Expression{
				{Name: "RESPONSES_SINCE_BY_KEY", Data: []types.ExpressionArg{
					{DType: "num", Num: float64(time.Now().Add(-20 * time.Hour * 24).Unix())},
					{Str: "s2"},
				}},
			},
		},
		PrefillRules: []types.Expression{
			{
				Name: "GET_LAST_SURVEY_ITEM", Data: []types.ExpressionArg{
					{Str: "s1"},
					{Str: "s1.1"},
				},
			},
			{
				Name: "GET_LAST_SURVEY_ITEM", Data: []types.ExpressionArg{
					{Str: "s2"},
					{Str: "s2.2"},
				},
			},
			{
				Name: "GET_LAST_SURVEY_ITEM", Data: []types.ExpressionArg{
					{Str: "s2"},
					{Str: "s2.4"},
				},
			},
		},
	}

	_, err = testStudyDBService.SaveSurvey(testInstanceID, testStudyKey, testSurvey)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}

	t.Run("with missing request", func(t *testing.T) {
		_, err := s.GetAssignedSurvey(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty request", func(t *testing.T) {
		_, err := s.GetAssignedSurvey(context.Background(), &api.SurveyReferenceRequest{})
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("wrong study key", func(t *testing.T) {
		_, err := s.GetAssignedSurvey(context.Background(), &api.SurveyReferenceRequest{
			Token:     &api.TokenInfos{Id: testUserID, InstanceId: testInstanceID},
			StudyKey:  "wrong",
			SurveyKey: "t1",
		})
		ok, msg := shouldHaveGrpcErrorStatus(err, "")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("correct values", func(t *testing.T) {
		resp, err := s.GetAssignedSurvey(context.Background(), &api.SurveyReferenceRequest{
			Token:     &api.TokenInfos{Id: testUserID, InstanceId: testInstanceID},
			StudyKey:  testStudyKey,
			SurveyKey: "t1",
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if resp.Context.Mode != "test" {
			t.Error("wrong mode")
		}
		if resp.Survey.Current.SurveyDefinition.Key != "t1" {
			t.Error("wrong survey key")
		}
	})
}

func TestSubmitStatusReportEndpoint(t *testing.T) {
	s := studyServiceServer{
		globalDBService:   testGlobalDBService,
		studyDBservice:    testStudyDBService,
		StudyGlobalSecret: "globsecretfortest1234",
	}

	studies := []types.Study{
		{
			Status:    "active",
			Key:       "studyfor_submitstatus1",
			SecretKey: "testsecret",
		},
		{
			Status:    "active",
			Key:       "studyfor_submitstatus2",
			SecretKey: "testsecret2",
		},
		{
			Status:    "active",
			Key:       "studyfor_submitstatus3",
			SecretKey: "testsecret3",
		},
	}

	for _, study := range studies {
		_, err := testStudyDBService.CreateStudy(testInstanceID, study)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
	}

	testUserID := "234234laaabbb3423aa"

	pid1, err := s.userIDToParticipantID(testInstanceID, "studyfor_submitstatus1", testUserID)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	pid2, err := s.userIDToParticipantID(testInstanceID, "studyfor_submitstatus2", testUserID)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	pState1 := types.ParticipantState{
		ParticipantID: pid1,
		StudyStatus:   "active",
		AssignedSurveys: []types.AssignedSurvey{
			{SurveyKey: "s1"},
		},
	}
	pState2 := types.ParticipantState{
		ParticipantID: pid2,
		StudyStatus:   "paused",
		AssignedSurveys: []types.AssignedSurvey{
			{SurveyKey: "s2"},
		},
	}

	_, err = testStudyDBService.SaveParticipantState(testInstanceID, "studyfor_submitstatus1", pState1)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	_, err = testStudyDBService.SaveParticipantState(testInstanceID, "studyfor_submitstatus2", pState2)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}

	t.Run("with missing request", func(t *testing.T) {
		_, err := s.SubmitStatusReport(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty request", func(t *testing.T) {
		_, err := s.SubmitStatusReport(context.Background(), &api.StatusReportRequest{})
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("correct values", func(t *testing.T) {
		resp, err := s.SubmitStatusReport(context.Background(), &api.StatusReportRequest{
			Token: &api.TokenInfos{
				Id:         testUserID,
				InstanceId: testInstanceID,
				ProfilId:   testUserID,
			},
			StatusSurvey: &api.SurveyResponse{
				Key:           "t1",
				ParticipantId: pid1,
				Responses: []*api.SurveyItemResponse{
					{Key: "1"},
				},
			},
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(resp.Surveys) != 1 {
			t.Errorf("unexpected number of surveys: %d", len(resp.Surveys))
			return
		}
		if resp.Surveys[0].SurveyKey != "s1" {
			t.Errorf("unexpected survey key: %s", resp.Surveys[0].SurveyKey)
		}
	})
}

func TestSubmitResponseEndpoint(t *testing.T) {
	s := studyServiceServer{
		globalDBService:   testGlobalDBService,
		studyDBservice:    testStudyDBService,
		StudyGlobalSecret: "globsecretfortest1234",
	}

	studies := []types.Study{
		{
			Status:    "active",
			Key:       "studyfor_submitsurvey1",
			SecretKey: "testsecret",
		},
		{
			Status:    "active",
			Key:       "studyfor_submitsurvey2",
			SecretKey: "testsecret2",
		},
		{
			Status:    "active",
			Key:       "studyfor_submitsurvey3",
			SecretKey: "testsecret3",
		},
	}

	for _, study := range studies {
		_, err := testStudyDBService.CreateStudy(testInstanceID, study)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
	}

	testUserID := "234234laaabbb3423"

	pid1, err := s.userIDToParticipantID(testInstanceID, "studyfor_submitsurvey1", testUserID)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	pid2, err := s.userIDToParticipantID(testInstanceID, "studyfor_submitsurvey2", testUserID)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	pState1 := types.ParticipantState{
		ParticipantID: pid1,
		StudyStatus:   "active",
		AssignedSurveys: []types.AssignedSurvey{
			{SurveyKey: "s1"},
		},
	}
	pState2 := types.ParticipantState{
		ParticipantID: pid2,
		StudyStatus:   "paused",
		AssignedSurveys: []types.AssignedSurvey{
			{SurveyKey: "s2"},
		},
	}

	_, err = testStudyDBService.SaveParticipantState(testInstanceID, "studyfor_submitsurvey1", pState1)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	_, err = testStudyDBService.SaveParticipantState(testInstanceID, "studyfor_submitsurvey2", pState2)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}

	t.Run("with missing request", func(t *testing.T) {
		_, err := s.SubmitResponse(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty request", func(t *testing.T) {
		_, err := s.SubmitResponse(context.Background(), &api.SubmitResponseReq{})
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	survResp := api.SurveyResponse{
		Key:           "sKey",
		ParticipantId: pid1,
		Responses: []*api.SurveyItemResponse{
			{Key: "1"},
		},
	}

	t.Run("wrong study key", func(t *testing.T) {
		_, err := s.SubmitResponse(context.Background(), &api.SubmitResponseReq{
			Token:    &api.TokenInfos{Id: testUserID, InstanceId: testInstanceID},
			StudyKey: "wrong_study",
			Response: &survResp,
		})
		ok, msg := shouldHaveGrpcErrorStatus(err, "")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("correct values", func(t *testing.T) {
		_, err := s.SubmitResponse(context.Background(), &api.SubmitResponseReq{
			Token:    &api.TokenInfos{Id: testUserID, InstanceId: testInstanceID, ProfilId: testUserID},
			StudyKey: studies[0].Key,
			Response: &survResp,
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
	})
}

func TestResolveContextRules(t *testing.T) {
	s := studyServiceServer{
		globalDBService:   testGlobalDBService,
		studyDBservice:    testStudyDBService,
		StudyGlobalSecret: "globsecretfortest1234",
	}
	testStudyKey := "teststudy_forresolvecontext"
	testUserID := "234234laaabbb3423"

	pid1, err := s.userIDToParticipantID(testInstanceID, "studyfor_submitsurvey1", testUserID)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	surveyResps := []types.SurveyResponse{
		// mix participants and order for submittedAt
		{Key: "s2", ParticipantID: pid1, SubmittedAt: time.Now().Add(-32 * time.Hour * 24).Unix()},
		{Key: "s1", ParticipantID: "u2", SubmittedAt: time.Now().Add(-29 * time.Hour * 24).Unix()},
		{Key: "s1", ParticipantID: pid1, SubmittedAt: time.Now().Add(-23 * time.Hour * 24).Unix()},
		{Key: "s1", ParticipantID: pid1, SubmittedAt: time.Now().Add(-6 * time.Hour * 24).Unix()},
		{Key: "s2", ParticipantID: pid1, SubmittedAt: time.Now().Add(-5 * time.Hour * 24).Unix()},
		{Key: "s1", ParticipantID: "u2", SubmittedAt: time.Now().Add(-6 * time.Hour * 24).Unix()},
		{Key: "s2", ParticipantID: "u2", SubmittedAt: time.Now().Add(-7 * time.Hour * 24).Unix()},
		{Key: "s1", ParticipantID: pid1, SubmittedAt: time.Now().Add(-15 * time.Hour * 24).Unix()},
		{Key: "s2", ParticipantID: pid1, SubmittedAt: time.Now().Add(-14 * time.Hour * 24).Unix()},
	}
	for _, sr := range surveyResps {
		err := testStudyDBService.AddSurveyResponse(testInstanceID, testStudyKey, sr)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
	}

	t.Run("resolve with nil", func(t *testing.T) {
		sCtx, err := s.resolveContextRules(testInstanceID, testStudyKey, pid1, nil)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if sCtx.Mode != "" {
			t.Errorf("unexpected mode: %s", sCtx.Mode)
		}
	})

	t.Run("resolve mode string arg", func(t *testing.T) {
		testRules := types.SurveyContextDef{
			Mode: &types.ExpressionArg{Str: "test"},
		}
		sCtx, err := s.resolveContextRules(testInstanceID, testStudyKey, pid1, &testRules)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if sCtx.Mode != "test" {
			t.Errorf("unexpected mode: %s", sCtx.Mode)
		}
	})

	t.Run("find old responses since", func(t *testing.T) {
		testRules := types.SurveyContextDef{
			PreviousResponses: []types.Expression{
				{Name: "RESPONSES_SINCE_BY_KEY", Data: []types.ExpressionArg{
					{DType: "num", Num: float64(time.Now().Add(-20 * time.Hour * 24).Unix())},
					{Str: "s2"},
				}},
			},
		}
		sCtx, err := s.resolveContextRules(testInstanceID, testStudyKey, pid1, &testRules)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(sCtx.PreviousResponses) != 2 {
			t.Errorf("unexpected number of surveys: %d", len(sCtx.PreviousResponses))
		}
	})

	t.Run("find all old responses ", func(t *testing.T) {
		testRules := types.SurveyContextDef{
			PreviousResponses: []types.Expression{
				{Name: "ALL_RESPONSES_SINCE", Data: []types.ExpressionArg{
					{DType: "num", Num: float64(time.Now().Add(-20 * time.Hour * 24).Unix())},
				}},
			},
		}
		sCtx, err := s.resolveContextRules(testInstanceID, testStudyKey, pid1, &testRules)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(sCtx.PreviousResponses) != 4 {
			t.Errorf("unexpected number of surveys: %d", len(sCtx.PreviousResponses))
		}
	})

	t.Run("find old responses by key", func(t *testing.T) {
		testRules := types.SurveyContextDef{
			PreviousResponses: []types.Expression{
				{Name: "LAST_RESPONSES_BY_KEY", Data: []types.ExpressionArg{
					{Str: "s1"},
					{DType: "num", Num: 1},
				}},
			},
		}
		sCtx, err := s.resolveContextRules(testInstanceID, testStudyKey, pid1, &testRules)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(sCtx.PreviousResponses) != 1 {
			t.Errorf("unexpected number of surveys: %d", len(sCtx.PreviousResponses))
			return
		}
		if sCtx.PreviousResponses[0].Key != "s1" {
			t.Errorf("unexpected survey key: %s", sCtx.PreviousResponses[0].Key)
		}
	})
}

func TestResolvePrefillRules(t *testing.T) {
	s := studyServiceServer{
		globalDBService:   testGlobalDBService,
		studyDBservice:    testStudyDBService,
		StudyGlobalSecret: "globsecretfortest1234",
	}
	testStudyKey := "teststudy_forresolveprefills"
	testUserID := "234234laaabbb3423"

	pid1, err := s.userIDToParticipantID(testInstanceID, "studyfor_submitsurvey1", testUserID)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	surveyResps := []types.SurveyResponse{
		// mix participants and order for submittedAt
		{Key: "s1", ParticipantID: pid1, SubmittedAt: time.Now().Add(-6 * time.Hour * 24).Unix(), Responses: []types.SurveyItemResponse{
			{Key: "s1.1"},
			{Key: "s1.2"},
			{Key: "s1.3"},
		}},
		{Key: "s2", ParticipantID: pid1, SubmittedAt: time.Now().Add(-5 * time.Hour * 24).Unix(), Responses: []types.SurveyItemResponse{
			{Key: "s2.1"},
			{Key: "s2.2"},
			{Key: "s2.3"},
		}},
		{Key: "s1", ParticipantID: pid1, SubmittedAt: time.Now().Add(-15 * time.Hour * 24).Unix(), Responses: []types.SurveyItemResponse{

			{Key: "s1.2"},
			{Key: "s1.3"},
		}},
		{Key: "s2", ParticipantID: pid1, SubmittedAt: time.Now().Add(-14 * time.Hour * 24).Unix(), Responses: []types.SurveyItemResponse{
			{Key: "s2.1"},
			{Key: "s2.2"},
			{Key: "s2.3"},
		}},
	}
	for _, sr := range surveyResps {
		err := testStudyDBService.AddSurveyResponse(testInstanceID, testStudyKey, sr)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
	}

	t.Run("find last survey by type and extract items", func(t *testing.T) {
		rules := []types.Expression{
			{
				Name: "GET_LAST_SURVEY_ITEM", Data: []types.ExpressionArg{
					{Str: "s1"},
					{Str: "s1.1"},
				},
			},
			{
				Name: "GET_LAST_SURVEY_ITEM", Data: []types.ExpressionArg{
					{Str: "s2"},
					{Str: "s2.2"},
				},
			},
			{
				Name: "GET_LAST_SURVEY_ITEM", Data: []types.ExpressionArg{
					{Str: "s2"},
					{Str: "s2.4"},
				},
			},
		}
		prefill, err := s.resolvePrefillRules(testInstanceID, testStudyKey, pid1, rules)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(prefill.Responses) != 2 {
			t.Errorf("unexpected number of responses: %d", len(prefill.Responses))
			return
		}
		if prefill.Responses[0].Key != "s1.1" || prefill.Responses[1].Key != "s2.2" {
			log.Println(prefill)
			t.Error("unexpected responses")
		}
	})
}

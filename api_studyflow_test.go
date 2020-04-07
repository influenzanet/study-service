package main

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/influenzanet/study-service/api"
	"github.com/influenzanet/study-service/models"
)

func TestCheckIfParticipantExists(t *testing.T) {
	// Test setup
	testStudyKey := "teststudy_checkifparticipantexists"

	pStates := []models.ParticipantState{
		models.ParticipantState{
			ParticipantID: "1",
			StudyStatus:   "active",
		},
	}

	for _, ps := range pStates {
		_, err := saveParticipantStateDB(testInstanceID, testStudyKey, ps)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
	}

	// Tests
	t.Run("with existing participant", func(t *testing.T) {
		if !checkIfParticipantExists(testInstanceID, testStudyKey, "1") {
			t.Error("should be true if participant exists")
		}
	})

	t.Run("with not existing participant", func(t *testing.T) {
		if checkIfParticipantExists(testInstanceID, testStudyKey, "2") {
			t.Error("should be false if participant does not exist")
		}
	})
}

func TestGetAndPerformStudyRules(t *testing.T) {
	testStudy := models.Study{
		Key:       "studytocheckifrulesareworking",
		SecretKey: "testsecret",
		Rules: []models.Expression{
			models.Expression{
				Name: "IFTHEN",
				Data: []models.ExpressionArg{
					models.ExpressionArg{
						DType: "exp",
						Exp: models.Expression{
							Name: "checkEventType",
							Data: []models.ExpressionArg{
								models.ExpressionArg{Str: "ENTER"},
							},
						},
					},
					models.ExpressionArg{
						DType: "exp",
						Exp: models.Expression{
							Name: "UPDATE_FLAG",
							Data: []models.ExpressionArg{
								models.ExpressionArg{Str: "testKey"},
								models.ExpressionArg{Str: "testValue"},
							},
						},
					},
				},
			},
			models.Expression{
				Name: "IFTHEN",
				Data: []models.ExpressionArg{
					models.ExpressionArg{
						DType: "exp",
						Exp: models.Expression{
							Name: "checkEventType",
							Data: []models.ExpressionArg{
								models.ExpressionArg{Str: "SUBMIT"},
							},
						},
					},
					models.ExpressionArg{
						DType: "exp",
						Exp: models.Expression{
							Name: "UPDATE_FLAG",
							Data: []models.ExpressionArg{
								models.ExpressionArg{Str: "testKey"},
								models.ExpressionArg{Str: "testValue2"},
							},
						},
					},
				},
			},
		},
	}

	testStudy, err := createStudyInDB(testInstanceID, testStudy)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}

	pState := models.ParticipantState{
		ParticipantID: "1",
	}

	t.Run("ENTER event", func(t *testing.T) {
		testEvent := models.StudyEvent{
			Type: "ENTER",
		}

		pState, err = getAndPerformStudyRules(testInstanceID, testStudy.Key, pState, testEvent)
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
		testEvent := models.StudyEvent{
			Type: "SUBMIT",
			Response: models.SurveyResponse{
				Key: "testsurvey",
			},
		}

		pState, err = getAndPerformStudyRules(testInstanceID, testStudy.Key, pState, testEvent)
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
	s := studyServiceServer{}

	testStudy := models.Study{
		Key:       "studyfortestingenterstudy",
		SecretKey: "testsecret",
		Rules: []models.Expression{
			models.Expression{
				Name: "IFTHEN",
				Data: []models.ExpressionArg{
					models.ExpressionArg{
						DType: "exp",
						Exp: models.Expression{
							Name: "checkEventType",
							Data: []models.ExpressionArg{
								models.ExpressionArg{Str: "ENTER"},
							},
						},
					},
					models.ExpressionArg{
						DType: "exp",
						Exp: models.Expression{
							Name: "ADD_NEW_SURVEY",
							Data: []models.ExpressionArg{
								models.ExpressionArg{Str: "testsurvey"},
								models.ExpressionArg{DType: "num", Num: 0},
								models.ExpressionArg{DType: "num", Num: 0},
							},
						},
					},
				},
			},
			models.Expression{
				Name: "IFTHEN",
				Data: []models.ExpressionArg{
					models.ExpressionArg{
						DType: "exp",
						Exp: models.Expression{
							Name: "checkEventType",
							Data: []models.ExpressionArg{
								models.ExpressionArg{Str: "SUBMIT"},
							},
						},
					},
					models.ExpressionArg{
						DType: "exp",
						Exp: models.Expression{
							Name: "UPDATE_FLAG",
							Data: []models.ExpressionArg{
								models.ExpressionArg{Str: "testKey"},
								models.ExpressionArg{Str: "testValue2"},
							},
						},
					},
				},
			},
		},
	}

	testStudy, err := createStudyInDB(testInstanceID, testStudy)
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
	s := studyServiceServer{}

	studies := []models.Study{
		models.Study{
			Status:    "active",
			Key:       "studyforassignedsurvey1",
			SecretKey: "testsecret",
		},
		models.Study{
			Status:    "active",
			Key:       "studyforassignedsurvey2",
			SecretKey: "testsecret2",
		},
		models.Study{
			Status:    "active",
			Key:       "studyforassignedsurvey3",
			SecretKey: "testsecret3",
		},
	}

	for _, study := range studies {
		_, err := createStudyInDB(testInstanceID, study)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
	}

	testUserID := "234234laaabbb3423"

	pid1, err := userIDToParticipantID(testInstanceID, "studyforassignedsurvey1", testUserID)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	pid2, err := userIDToParticipantID(testInstanceID, "studyforassignedsurvey2", testUserID)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	pState1 := models.ParticipantState{
		ParticipantID: pid1,
		StudyStatus:   "active",
		AssignedSurveys: []models.AssignedSurvey{
			models.AssignedSurvey{SurveyKey: "s1"},
		},
	}
	pState2 := models.ParticipantState{
		ParticipantID: pid2,
		StudyStatus:   "active",
		AssignedSurveys: []models.AssignedSurvey{
			models.AssignedSurvey{SurveyKey: "s1"},
		},
	}

	_, err = saveParticipantStateDB(testInstanceID, "studyforassignedsurvey1", pState1)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	_, err = saveParticipantStateDB(testInstanceID, "studyforassignedsurvey2", pState2)
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
	s := studyServiceServer{}
	t.Run("with missing request", func(t *testing.T) {
		_, err := s.GetAssignedSurvey(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty request", func(t *testing.T) {
		_, err := s.GetAssignedSurvey(context.Background(), &api.GetSurveyRequest{})
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("wrong study key", func(t *testing.T) {
		t.Error("test unimplemented")
	})

	t.Run("correct values", func(t *testing.T) {
		t.Error("test unimplemented")
	})
}

func TestSubmitStatusReportEndpoint(t *testing.T) {
	s := studyServiceServer{}

	studies := []models.Study{
		models.Study{
			Status:    "active",
			Key:       "studyfor_submitstatus1",
			SecretKey: "testsecret",
		},
		models.Study{
			Status:    "active",
			Key:       "studyfor_submitstatus2",
			SecretKey: "testsecret2",
		},
		models.Study{
			Status:    "active",
			Key:       "studyfor_submitstatus3",
			SecretKey: "testsecret3",
		},
	}

	for _, study := range studies {
		_, err := createStudyInDB(testInstanceID, study)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
	}

	testUserID := "234234laaabbb3423aa"

	pid1, err := userIDToParticipantID(testInstanceID, "studyfor_submitstatus1", testUserID)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	pid2, err := userIDToParticipantID(testInstanceID, "studyfor_submitstatus2", testUserID)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	pState1 := models.ParticipantState{
		ParticipantID: pid1,
		StudyStatus:   "active",
		AssignedSurveys: []models.AssignedSurvey{
			models.AssignedSurvey{SurveyKey: "s1"},
		},
	}
	pState2 := models.ParticipantState{
		ParticipantID: pid2,
		StudyStatus:   "paused",
		AssignedSurveys: []models.AssignedSurvey{
			models.AssignedSurvey{SurveyKey: "s2"},
		},
	}

	_, err = saveParticipantStateDB(testInstanceID, "studyfor_submitstatus1", pState1)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	_, err = saveParticipantStateDB(testInstanceID, "studyfor_submitstatus2", pState2)
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
			},
			StatusSurvey: &api.SurveyResponse{
				Key: "t1",
				Responses: []*api.SurveyItemResponse{
					&api.SurveyItemResponse{Key: "1"},
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
	s := studyServiceServer{}

	studies := []models.Study{
		models.Study{
			Status:    "active",
			Key:       "studyfor_submitsurvey1",
			SecretKey: "testsecret",
		},
		models.Study{
			Status:    "active",
			Key:       "studyfor_submitsurvey2",
			SecretKey: "testsecret2",
		},
		models.Study{
			Status:    "active",
			Key:       "studyfor_submitsurvey3",
			SecretKey: "testsecret3",
		},
	}

	for _, study := range studies {
		_, err := createStudyInDB(testInstanceID, study)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
	}

	testUserID := "234234laaabbb3423"

	pid1, err := userIDToParticipantID(testInstanceID, "studyfor_submitsurvey1", testUserID)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	pid2, err := userIDToParticipantID(testInstanceID, "studyfor_submitsurvey2", testUserID)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	pState1 := models.ParticipantState{
		ParticipantID: pid1,
		StudyStatus:   "active",
		AssignedSurveys: []models.AssignedSurvey{
			models.AssignedSurvey{SurveyKey: "s1"},
		},
	}
	pState2 := models.ParticipantState{
		ParticipantID: pid2,
		StudyStatus:   "paused",
		AssignedSurveys: []models.AssignedSurvey{
			models.AssignedSurvey{SurveyKey: "s2"},
		},
	}

	_, err = saveParticipantStateDB(testInstanceID, "studyfor_submitsurvey1", pState1)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	_, err = saveParticipantStateDB(testInstanceID, "studyfor_submitsurvey2", pState2)
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
		Key: "sKey",
		Responses: []*api.SurveyItemResponse{
			&api.SurveyItemResponse{Key: "1"},
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
			Token:    &api.TokenInfos{Id: testUserID, InstanceId: testInstanceID},
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
	testStudyKey := "teststudy_forresolvecontext"
	testUserID := "234234laaabbb3423"

	pid1, err := userIDToParticipantID(testInstanceID, "studyfor_submitsurvey1", testUserID)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	surveyResps := []models.SurveyResponse{
		// mix participants and order for submittedAt
		models.SurveyResponse{Key: "s1", SubmittedFor: pid1, SubmittedAt: time.Now().Add(-30 * time.Hour * 24).Unix()},
		models.SurveyResponse{Key: "s2", SubmittedFor: pid1, SubmittedAt: time.Now().Add(-32 * time.Hour * 24).Unix()},
		models.SurveyResponse{Key: "s1", SubmittedFor: "u2", SubmittedAt: time.Now().Add(-29 * time.Hour * 24).Unix()},
		models.SurveyResponse{Key: "s1", SubmittedFor: pid1, SubmittedAt: time.Now().Add(-23 * time.Hour * 24).Unix()},
		models.SurveyResponse{Key: "s1", SubmittedFor: pid1, SubmittedAt: time.Now().Add(-6 * time.Hour * 24).Unix()},
		models.SurveyResponse{Key: "s2", SubmittedFor: pid1, SubmittedAt: time.Now().Add(-5 * time.Hour * 24).Unix()},
		models.SurveyResponse{Key: "s1", SubmittedFor: "u2", SubmittedAt: time.Now().Add(-6 * time.Hour * 24).Unix()},
		models.SurveyResponse{Key: "s2", SubmittedFor: "u2", SubmittedAt: time.Now().Add(-7 * time.Hour * 24).Unix()},
		models.SurveyResponse{Key: "s1", SubmittedFor: pid1, SubmittedAt: time.Now().Add(-15 * time.Hour * 24).Unix()},
		models.SurveyResponse{Key: "s2", SubmittedFor: pid1, SubmittedAt: time.Now().Add(-14 * time.Hour * 24).Unix()},
	}
	for _, sr := range surveyResps {
		err := addSurveyResponseToDB(testInstanceID, testStudyKey, sr)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
	}

	t.Run("resolve mode string arg", func(t *testing.T) {
		testRules := models.SurveyContextDef{
			Mode: models.ExpressionArg{Str: "test"},
		}
		sCtx, err := resolveContextRules(testInstanceID, testStudyKey, pid1, testRules)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if sCtx.Mode != "test" {
			t.Errorf("unexpected mode: %s", sCtx.Mode)
		}
	})

	t.Run("find old responses since", func(t *testing.T) {
		testRules := models.SurveyContextDef{
			PreviousResponses: []models.Expression{
				models.Expression{Name: "RESPONSES_SINCE_BY_KEY", Data: []models.ExpressionArg{
					models.ExpressionArg{DType: "num", Num: float64(time.Now().Add(-20 * time.Hour * 24).Unix())},
					models.ExpressionArg{Str: "s2"},
				}},
			},
		}
		sCtx, err := resolveContextRules(testInstanceID, testStudyKey, pid1, testRules)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(sCtx.PreviousResponses) != 2 {
			t.Errorf("unexpected number of surveys: %d", len(sCtx.PreviousResponses))
		}
	})

	t.Run("find all old responses ", func(t *testing.T) {
		testRules := models.SurveyContextDef{
			PreviousResponses: []models.Expression{
				models.Expression{Name: "ALL_RESPONSES_SINCE", Data: []models.ExpressionArg{
					models.ExpressionArg{DType: "num", Num: float64(time.Now().Add(-20 * time.Hour * 24).Unix())},
				}},
			},
		}
		sCtx, err := resolveContextRules(testInstanceID, testStudyKey, pid1, testRules)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(sCtx.PreviousResponses) != 4 {
			t.Errorf("unexpected number of surveys: %d", len(sCtx.PreviousResponses))
		}
	})

	t.Run("find old responses by key", func(t *testing.T) {
		testRules := models.SurveyContextDef{
			PreviousResponses: []models.Expression{
				models.Expression{Name: "LAST_RESPONSES_BY_KEY", Data: []models.ExpressionArg{
					models.ExpressionArg{Str: "s1"},
					models.ExpressionArg{DType: "num", Num: 1},
				}},
			},
		}
		sCtx, err := resolveContextRules(testInstanceID, testStudyKey, pid1, testRules)
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
	testStudyKey := "teststudy_forresolveprefills"
	testUserID := "234234laaabbb3423"

	pid1, err := userIDToParticipantID(testInstanceID, "studyfor_submitsurvey1", testUserID)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	surveyResps := []models.SurveyResponse{
		// mix participants and order for submittedAt
		models.SurveyResponse{Key: "s1", SubmittedFor: pid1, SubmittedAt: time.Now().Add(-6 * time.Hour * 24).Unix(), Responses: []models.SurveyItemResponse{
			models.SurveyItemResponse{Key: "s1.1"},
			models.SurveyItemResponse{Key: "s1.2"},
			models.SurveyItemResponse{Key: "s1.3"},
		}},
		models.SurveyResponse{Key: "s2", SubmittedFor: pid1, SubmittedAt: time.Now().Add(-5 * time.Hour * 24).Unix(), Responses: []models.SurveyItemResponse{
			models.SurveyItemResponse{Key: "s2.1"},
			models.SurveyItemResponse{Key: "s2.2"},
			models.SurveyItemResponse{Key: "s2.3"},
		}},
		models.SurveyResponse{Key: "s1", SubmittedFor: pid1, SubmittedAt: time.Now().Add(-15 * time.Hour * 24).Unix(), Responses: []models.SurveyItemResponse{
			models.SurveyItemResponse{Key: "s1.1"},
			models.SurveyItemResponse{Key: "s1.2"},
			models.SurveyItemResponse{Key: "s1.3"},
		}},
		models.SurveyResponse{Key: "s2", SubmittedFor: pid1, SubmittedAt: time.Now().Add(-14 * time.Hour * 24).Unix(), Responses: []models.SurveyItemResponse{
			models.SurveyItemResponse{Key: "s2.1"},
			models.SurveyItemResponse{Key: "s2.2"},
			models.SurveyItemResponse{Key: "s2.3"},
		}},
	}
	for _, sr := range surveyResps {
		err := addSurveyResponseToDB(testInstanceID, testStudyKey, sr)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
	}

	t.Run("find last survey by type and extract items", func(t *testing.T) {
		rules := []models.Expression{
			models.Expression{
				Name: "GET_LAST_SURVEY_ITEM", Data: []models.ExpressionArg{
					models.ExpressionArg{Str: "s1"},
					models.ExpressionArg{Str: "s1.1"},
				},
			},
			models.Expression{
				Name: "GET_LAST_SURVEY_ITEM", Data: []models.ExpressionArg{
					models.ExpressionArg{Str: "s2"},
					models.ExpressionArg{Str: "s2.2"},
				},
			},
			models.Expression{
				Name: "GET_LAST_SURVEY_ITEM", Data: []models.ExpressionArg{
					models.ExpressionArg{Str: "s2"},
					models.ExpressionArg{Str: "s2.4"},
				},
			},
		}
		prefill, err := resolvePrefillRules(testInstanceID, testStudyKey, pid1, rules)
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

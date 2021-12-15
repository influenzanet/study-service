package service

import (
	"context"
	"testing"
	"time"

	"github.com/influenzanet/go-utils/pkg/api_types"
	"github.com/influenzanet/study-service/pkg/api"
	"github.com/influenzanet/study-service/pkg/types"
	"github.com/influenzanet/study-service/pkg/utils"
)

func TestGetStudiesForUserEndpoint(t *testing.T) {
	s := studyServiceServer{
		globalDBService:   testGlobalDBService,
		studyDBservice:    testStudyDBService,
		StudyGlobalSecret: "globsecretfortest1234",
	}
	testStudies := []types.Study{
		{
			Status:    "active",
			Key:       "studyfor_getstudiesforuser_1",
			SecretKey: "testsecret",
		}, {
			Status:    "archived",
			Key:       "studyfor_getstudiesforuser_2",
			SecretKey: "testsecret2",
		}, {
			Status:    "active",
			Key:       "studyfor_getstudiesforuser_3",
			SecretKey: "testsecret3",
		}, {
			Status:    "archived",
			Key:       "studyfor_getstudiesforuser_4",
			SecretKey: "testsecret4",
		},
	}

	for _, study := range testStudies {
		_, err := testStudyDBService.CreateStudy(testInstanceID, study)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
	}

	testProfileID1 := "234234laaabbb3423_for_getstudies"
	testProfileID2 := "234234laaabbb3423_for_getstudies2"

	pid1, err := s.profileIDToParticipantID(testInstanceID, testStudies[0].Key, testProfileID1)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	pid2, err := s.profileIDToParticipantID(testInstanceID, testStudies[1].Key, testProfileID1)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	pid3, err := s.profileIDToParticipantID(testInstanceID, testStudies[2].Key, testProfileID1)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	pid4, err := s.profileIDToParticipantID(testInstanceID, testStudies[3].Key, testProfileID2)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	pState1 := types.ParticipantState{
		ParticipantID: pid1,
		StudyStatus:   types.PARTICIPANT_STUDY_STATUS_ACTIVE,
		AssignedSurveys: []types.AssignedSurvey{
			{SurveyKey: "s1"},
		},
	}
	pState2 := types.ParticipantState{
		ParticipantID: pid2,
		StudyStatus:   types.PARTICIPANT_STUDY_STATUS_ACTIVE,
		AssignedSurveys: []types.AssignedSurvey{
			{SurveyKey: "s2"},
		},
	}
	pState3 := types.ParticipantState{
		ParticipantID: pid3,
		StudyStatus:   types.PARTICIPANT_STUDY_STATUS_EXITED,
	}
	pState4 := types.ParticipantState{
		ParticipantID: pid4,
		StudyStatus:   types.PARTICIPANT_STUDY_STATUS_ACTIVE,
		AssignedSurveys: []types.AssignedSurvey{
			{SurveyKey: "s2"},
		},
	}

	_, err = s.studyDBservice.SaveParticipantState(testInstanceID, testStudies[0].Key, pState1)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	_, err = s.studyDBservice.SaveParticipantState(testInstanceID, testStudies[1].Key, pState2)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	_, err = s.studyDBservice.SaveParticipantState(testInstanceID, testStudies[2].Key, pState3)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	_, err = s.studyDBservice.SaveParticipantState(testInstanceID, testStudies[3].Key, pState4)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}

	t.Run("with missing request", func(t *testing.T) {
		_, err := s.GetStudiesForUser(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty request", func(t *testing.T) {
		_, err := s.GetStudiesForUser(context.Background(), &api.GetStudiesForUserReq{})
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with valid request", func(t *testing.T) {
		resp, err := s.GetStudiesForUser(context.Background(), &api.GetStudiesForUserReq{
			Token: &api_types.TokenInfos{
				Id:              "userid",
				InstanceId:      testInstanceID,
				ProfilId:        testProfileID1,
				OtherProfileIds: []string{testProfileID2},
			},
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if len(resp.Studies) != 3 {
			t.Errorf("unexpected number of studies: %d", len(resp.Studies))
		}
	})
}

func TestGetActiveStudiesEndpoint(t *testing.T) {
	s := studyServiceServer{
		globalDBService:   testGlobalDBService,
		studyDBservice:    testStudyDBService,
		StudyGlobalSecret: "globsecretfortest1234",
	}
	testStudies := []types.Study{
		{
			Status:    "active",
			Key:       "studyfor_getactivestudies_1",
			SecretKey: "testsecret",
		}, {
			Status:    "archived",
			Key:       "studyfor_getactivestudies_2",
			SecretKey: "testsecret2",
		}, {
			Status:    "active",
			Key:       "studyfor_getactivestudies_3",
			SecretKey: "testsecret3",
		}, {
			Status:    "active",
			Key:       "studyfor_getactivestudies_4",
			SecretKey: "testsecret4",
		},
	}

	for _, study := range testStudies {
		_, err := testStudyDBService.CreateStudy(testInstanceID, study)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
	}

	t.Run("with missing request", func(t *testing.T) {
		_, err := s.GetActiveStudies(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty request", func(t *testing.T) {
		_, err := s.GetActiveStudies(context.Background(), &api_types.TokenInfos{})
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with valid request", func(t *testing.T) {
		resp, err := s.GetActiveStudies(context.Background(), &api_types.TokenInfos{
			Id:              "userid",
			InstanceId:      testInstanceID,
			ProfilId:        "testProfileID1",
			OtherProfileIds: []string{"testProfileID2"},
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if len(resp.Studies) < 3 {
			t.Errorf("unexpected number of studies: %d", len(resp.Studies))
		}
	})
}
func TestHasParticipantStateWithConditionEndpoint(t *testing.T) {
	s := studyServiceServer{
		globalDBService:   testGlobalDBService,
		studyDBservice:    testStudyDBService,
		StudyGlobalSecret: "globsecretfortest1234",
	}

	// create study for user in it
	testStudies := []types.Study{
		{
			Status:    "active",
			Key:       "studyfor_hasParticipantWithConditionStudies_1",
			SecretKey: "testsecret",
			Configs: types.StudyConfigs{
				IdMappingMethod: utils.ID_MAPPING_AESCTR,
			},
		},
	}

	for _, study := range testStudies {
		_, err := testStudyDBService.CreateStudy(testInstanceID, study)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
	}

	testProfileID1 := "234234laaabbb3423_for_hasProf_1"
	testProfileID2 := "234234laaabbb3423_for_hasProf_2"

	pid1, err := s.profileIDToParticipantID(testInstanceID, testStudies[0].Key, testProfileID1)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	pid2, err := s.profileIDToParticipantID(testInstanceID, testStudies[0].Key, testProfileID2)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}

	pState1 := types.ParticipantState{
		ParticipantID: pid1,
		StudyStatus:   types.PARTICIPANT_STUDY_STATUS_ACTIVE,
		LastSubmissions: map[string]int64{
			"s3": time.Now().Unix() - 20,
		},
		AssignedSurveys: []types.AssignedSurvey{
			{SurveyKey: "s1"},
		},
	}

	pState2 := types.ParticipantState{
		ParticipantID: pid2,
		StudyStatus:   types.PARTICIPANT_STUDY_STATUS_EXITED,
	}

	_, err = s.studyDBservice.SaveParticipantState(testInstanceID, testStudies[0].Key, pState1)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	_, err = s.studyDBservice.SaveParticipantState(testInstanceID, testStudies[0].Key, pState2)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}

	t.Run("with missing request", func(t *testing.T) {
		_, err := s.HasParticipantStateWithCondition(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty request", func(t *testing.T) {
		_, err := s.HasParticipantStateWithCondition(context.Background(), &api.ProfilesWithConditionReq{})
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with user profiles not in the study", func(t *testing.T) {
		_, err := s.HasParticipantStateWithCondition(context.Background(), &api.ProfilesWithConditionReq{
			ProfileIds: []string{"notthere1", "notthere2"},
			StudyKey:   testStudies[0].Key,
			InstanceId: testInstanceID,
			Condition:  &api.ExpressionArg{Dtype: "num", Data: &api.ExpressionArg_Num{Num: 1}},
		})
		ok, msg := shouldHaveGrpcErrorStatus(err, "no participant found")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with profiles not fulfilling condition", func(t *testing.T) {
		_, err := s.HasParticipantStateWithCondition(context.Background(), &api.ProfilesWithConditionReq{
			ProfileIds: []string{testProfileID2},
			StudyKey:   testStudies[0].Key,
			InstanceId: testInstanceID,
			Condition: &api.ExpressionArg{Dtype: "exp", Data: &api.ExpressionArg_Exp{Exp: &api.Expression{
				Name: "hasStudyStatus",
				Data: []*api.ExpressionArg{
					{Dtype: "str", Data: &api.ExpressionArg_Str{Str: types.PARTICIPANT_STUDY_STATUS_ACTIVE}},
				},
			}}},
		})
		ok, msg := shouldHaveGrpcErrorStatus(err, "no participant found")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with hardcoded condition", func(t *testing.T) {
		_, err := s.HasParticipantStateWithCondition(context.Background(), &api.ProfilesWithConditionReq{
			ProfileIds: []string{testProfileID1},
			StudyKey:   testStudies[0].Key,
			InstanceId: testInstanceID,
			Condition:  &api.ExpressionArg{Dtype: "num", Data: &api.ExpressionArg_Num{Num: 1}},
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
	})

	t.Run("with profiles fulfilling condition", func(t *testing.T) {
		_, err := s.HasParticipantStateWithCondition(context.Background(), &api.ProfilesWithConditionReq{
			ProfileIds: []string{testProfileID1},
			StudyKey:   testStudies[0].Key,
			InstanceId: testInstanceID,
			Condition: &api.ExpressionArg{Dtype: "exp", Data: &api.ExpressionArg_Exp{Exp: &api.Expression{
				Name: "lastSubmissionDateOlderThan",
				Data: []*api.ExpressionArg{
					{Dtype: "exp", Data: &api.ExpressionArg_Exp{Exp: &api.Expression{Name: "timestampWithOffset", Data: []*api.ExpressionArg{
						{Dtype: "num", Data: &api.ExpressionArg_Num{Num: 10}},
					}}}},
					{Dtype: "str", Data: &api.ExpressionArg_Str{Str: "s3"}},
				},
			}}},
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}

		_, err = s.HasParticipantStateWithCondition(context.Background(), &api.ProfilesWithConditionReq{
			ProfileIds: []string{testProfileID2},
			StudyKey:   testStudies[0].Key,
			InstanceId: testInstanceID,
			Condition: &api.ExpressionArg{Dtype: "exp", Data: &api.ExpressionArg_Exp{Exp: &api.Expression{
				Name: "hasStudyStatus",
				Data: []*api.ExpressionArg{
					{Dtype: "str", Data: &api.ExpressionArg_Str{Str: "exited"}},
				},
			}}},
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
	})
}

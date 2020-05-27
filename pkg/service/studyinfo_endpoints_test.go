package service

import (
	"context"
	"testing"

	"github.com/influenzanet/study-service/pkg/api"
	"github.com/influenzanet/study-service/pkg/types"
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
		StudyStatus:   "active",
		AssignedSurveys: []types.AssignedSurvey{
			{SurveyKey: "s1"},
		},
	}
	pState2 := types.ParticipantState{
		ParticipantID: pid2,
		StudyStatus:   "active",
		AssignedSurveys: []types.AssignedSurvey{
			{SurveyKey: "s2"},
		},
	}
	pState3 := types.ParticipantState{
		ParticipantID: pid3,
		StudyStatus:   "exited",
	}
	pState4 := types.ParticipantState{
		ParticipantID: pid4,
		StudyStatus:   "active",
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
			Token: &api.TokenInfos{
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
		_, err := s.GetActiveStudies(context.Background(), &api.TokenInfos{})
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with valid request", func(t *testing.T) {
		resp, err := s.GetActiveStudies(context.Background(), &api.TokenInfos{
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
	/*s := studyServiceServer{
		globalDBService:   testGlobalDBService,
		studyDBservice:    testStudyDBService,
		StudyGlobalSecret: "globsecretfortest1234",
	}*/
	// create study for user in it

	// test with nil
	// test with empty
	// test with user profiles not in the study
	// test with profiles not fulfilling condition
	// test with profiles mathing conditions
	t.Error("test unimplemented")
}

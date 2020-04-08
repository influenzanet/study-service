package main

import (
	"context"
	"testing"

	"github.com/influenzanet/study-service/api"
)

func TestCreateNewStudyEndpoint(t *testing.T) {
	s := studyServiceServer{}

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

func TestAddSurveyToStudyEndpoint(t *testing.T) {
	s := studyServiceServer{}

	t.Run("with missing request", func(t *testing.T) {
		_, err := s.AddSurveyToStudy(context.Background(), nil)
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with empty request", func(t *testing.T) {
		_, err := s.AddSurveyToStudy(context.Background(), &api.AddSurveyReq{})
		ok, msg := shouldHaveGrpcErrorStatus(err, "missing argument")
		if !ok {
			t.Error(msg)
		}
	})

	t.Run("with new survey to add", func(t *testing.T) {
		t.Error("test unimplemented")
	})
}

func TestGetStudySurveyInfosEndpoint(t *testing.T) {
	s := studyServiceServer{}

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
		t.Error("test unimplemented")
	})
}

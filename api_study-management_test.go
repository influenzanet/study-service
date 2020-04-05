package main

import (
		"context"
	"testing"

	"github.com/influenzanet/study-service/api"
)

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

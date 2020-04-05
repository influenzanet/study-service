package main

import (
	"context"
	"testing"

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
	// TODO: setup study rules
	t.Run("ENTER event", func(t *testing.T) {
		t.Error("test unimplemented")
	})
	t.Run("SUBMIT event", func(t *testing.T) {
		t.Error("test unimplemented")
	})
}

func TestEnterStudyEndpoint(t *testing.T) {
	s := studyServiceServer{}

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
		t.Error("test unimplemented")
	})

	t.Run("correct values", func(t *testing.T) {
		t.Error("test unimplemented")
	})

	t.Run("existing participant (user) id", func(t *testing.T) {
		t.Error("test unimplemented")
	})
}

func TestGetAssignedSurveysEndpoint(t *testing.T) {
	s := studyServiceServer{}
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
		t.Error("test unimplemented")
	})

	t.Run("correct values", func(t *testing.T) {
		t.Error("test unimplemented")
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

	t.Run("wrong study key", func(t *testing.T) {
		t.Error("test unimplemented")
	})

	t.Run("correct values", func(t *testing.T) {
		t.Error("test unimplemented")
	})
}

func TestSubmitResponseEndpoint(t *testing.T) {
	s := studyServiceServer{}
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

	t.Run("wrong study key", func(t *testing.T) {
		t.Error("test unimplemented")
	})

	t.Run("correct values", func(t *testing.T) {
		t.Error("test unimplemented")
	})
}

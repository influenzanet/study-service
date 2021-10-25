package studydb

import (
	"testing"

	"github.com/influenzanet/study-service/pkg/types"
)

func TestDbCreateStudyInfos(t *testing.T) {
	testStudy := types.Study{
		Key:       "testStudyKey1",
		SecretKey: "supersecret",
		Members: []types.StudyMember{
			{
				UserID: "testuser",
				Role:   "maintainer",
			},
		},
	}

	t.Run("Create study with not existing key", func(t *testing.T) {
		study, err := testDBService.CreateStudy(testInstanceID, testStudy)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if study.ID.IsZero() {
			t.Errorf("unexpected id: %s", study.ID.Hex())
		}
	})

	t.Run("Try to create study with existing key", func(t *testing.T) {
		_, err := testDBService.CreateStudy(testInstanceID, testStudy)
		if err == nil {
			t.Error("should return error")
		}
	})
}

func TestDbUpdateStudyInfos(t *testing.T) {
	testStudies := []types.Study{
		{Key: "test1", Status: "active", Members: []types.StudyMember{
			{
				UserID: "testuser",
				Role:   "maintainer",
			},
		}},
		{Key: "test2", Status: "active", Members: []types.StudyMember{
			{
				UserID: "testuser",
				Role:   "maintainer",
			},
		}},
	}

	for _, s := range testStudies {
		_, err := testDBService.CreateStudy(testInstanceID, s)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
	}

	t.Run("Try to update study key with existing key", func(t *testing.T) {
		if err := testDBService.UpdateStudyKey(testInstanceID, "test1", "test2"); err == nil {
			t.Error("should fail with error when key exists")
		}
	})

	t.Run("Try to update study key with not existing key", func(t *testing.T) {
		if err := testDBService.UpdateStudyKey(testInstanceID, "test1", "test3"); err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
	})

	t.Run("Update study status", func(t *testing.T) {
		if err := testDBService.UpdateStudyStatus(testInstanceID, "test1", "inactive"); err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
	})

	t.Run("Try to update other properties with ok key", func(t *testing.T) {
		testStudies[1].SecretKey = "343434"
		upd, err := testDBService.UpdateStudyInfo(testInstanceID, testStudies[1])
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if upd.SecretKey != "343434" {
			t.Errorf("unexpected value: %s, %s (have, want)", upd.SecretKey, "343434")
		}
	})

	t.Run("Try to update other properties with wrong key", func(t *testing.T) {
		testStudies[1].Key = "wrong"
		testStudies[1].SecretKey = "34343w4"
		_, err := testDBService.UpdateStudyInfo(testInstanceID, testStudies[1])
		if err == nil {
			t.Error("should return error since key not known")
			return
		}
	})
}

func TestDbGetStudyInfos(t *testing.T) {
	testStudies := []types.Study{
		{
			Key:       "testg1",
			SecretKey: "testsecret",
			Status:    "active",
			Members: []types.StudyMember{
				{
					UserID: "testuser",
					Role:   "maintainer",
				},
			},
			Rules: []types.Expression{
				{Name: "IFTHEN"}, // These here are not complete and won't be evaluated in this test
				{Name: "TEST"},
			},
			Configs: types.StudyConfigs{
				IdMappingMethod: "sha256",
			},
		},
		{Key: "testG2", SecretKey: "testsecret", Status: "inactive", Members: []types.StudyMember{
			{
				UserID: "testuser",
				Role:   "maintainer",
			},
		}, Configs: types.StudyConfigs{
			IdMappingMethod: "same",
		}},
	}

	for _, s := range testStudies {
		_, err := testDBService.CreateStudy(testInstanceID, s)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
	}

	t.Run("Get secret key", func(t *testing.T) {
		idMapping, secret, err := testDBService.GetStudySecretKey(testInstanceID, testStudies[0].Key)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if secret != testStudies[0].SecretKey {
			t.Errorf("unexpected value: %s, %s (have, want)", secret, testStudies[0].SecretKey)
		}
		if idMapping != testStudies[0].Configs.IdMappingMethod {
			t.Errorf("unexpected value: %s, %s (have, want)", idMapping, testStudies[0].Configs.IdMappingMethod)
		}
	})

	t.Run("Get members", func(t *testing.T) {
		members, err := testDBService.GetStudyMembers(testInstanceID, testStudies[0].Key)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(members) != 1 {
			t.Errorf("unexpected number of members: %d", len(members))
		}
	})

	t.Run("Get studies by status", func(t *testing.T) {
		studies, err := testDBService.GetStudiesByStatus(testInstanceID, "inactive", true)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(studies) != 1 {
			t.Errorf("unexpected number of studies: %d", len(studies))
			return
		}
		if studies[0].Status != "" {
			t.Error("should return only key and secretKey")
		}
	})

	t.Run("Get study rule", func(t *testing.T) {
		rules, err := testDBService.GetStudyRules(testInstanceID, testStudies[0].Key)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if len(rules) != 2 {
			t.Errorf("unexpected number of rules: %d", len(rules))
			return
		}
		if rules[0].Name != "IFTHEN" {
			t.Error("wrong expression")
		}
	})
}

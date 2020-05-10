package utils

import (
	"testing"

	"github.com/influenzanet/study-service/pkg/types"
)

func TestMembershipChecker(t *testing.T) {
	members := []types.StudyMember{
		{
			UserID: "testuser",
			Role:   "testrole",
		},
		{
			UserID: "testuser2",
			Role:   "testrole",
		},
	}
	t.Run("not existing member", func(t *testing.T) {
		val := CheckIfMember("testuser3", members, []string{})
		if val {
			t.Error("should be false")
		}
	})
	t.Run("existing member", func(t *testing.T) {
		val := CheckIfMember("testuser2", members, []string{})
		if !val {
			t.Error("should be true")
		}
	})
	t.Run("existing member with wrong role", func(t *testing.T) {
		val := CheckIfMember("testuser2", members, []string{"otherrole"})
		if val {
			t.Error("should be false")
		}
	})
}

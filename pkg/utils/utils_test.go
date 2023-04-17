package utils

import (
	"testing"

	"github.com/influenzanet/study-service/pkg/types"
)

func TestGenerateSurveyVersionID(t *testing.T) {
	t.Run("test id generation for uniqueness", func(t *testing.T) {
		oldVersions := []*types.Survey{}

		for i := 0; i < 100; i++ {
			id := GenerateSurveyVersionID(oldVersions)
			oldVersions = append(oldVersions, &types.Survey{VersionID: id})
		}

		for i, id_1 := range oldVersions {
			for j, id_2 := range oldVersions {
				if i != j && id_1.VersionID == id_2.VersionID {
					t.Errorf("duplicate key present: i: %d - %s j: %d - %s ", i, id_1.VersionID, j, id_2.VersionID)
				}
			}
		}
	})
}

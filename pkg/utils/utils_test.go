package utils

import (
	"testing"
	"time"
)

func TestGenerateSurveyVersionID(t *testing.T) {
	t.Run("test id generation for uniqueness", func(t *testing.T) {
		ids := []string{}

		for i := 0; i < 100; i++ {
			id := GenerateSurveyVersionID()
			ids = append(ids, id)
			time.Sleep(time.Millisecond)
		}

		for i, id_1 := range ids {
			for j, id_2 := range ids {
				if i != j && id_1 == id_2 {
					t.Errorf("duplicate key present: i: %d - %s j: %d - %s ", i, id_1, j, id_2)
				}
			}
		}
	})
}

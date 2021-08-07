package studydb

import (
	"testing"
	"time"

	"github.com/influenzanet/study-service/pkg/types"
)

func TestDbAddFileInfo(t *testing.T) {
	testStudy := "testosaveresponse"
	testFileInfo := types.FileInfo{
		ParticipantID: "testparticipantID",
		Status:        "Uploading",
		SubmittedAt:   time.Now().Unix(),
	}
	t.Run("saving file info", func(t *testing.T) {
		info, err := testDBService.SaveFileInfo(testInstanceID, testStudy, testFileInfo)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if info.ID.IsZero() || info.ID.Hex() == "" {
			t.Errorf("unexpected id: %v", info.ID)
		}
	})
}

package utils

import (
	"fmt"
	"time"

	"github.com/influenzanet/study-service/pkg/types"
)

func GenerateSurveyVersionID(oldVersions []*types.Survey) string {
	t := time.Now()

	date := t.Format("06-01")

	counter := 1
	newID := fmt.Sprintf("%s-%d", date, counter)
	for {
		idAlreadyPresent := false
		for _, v := range oldVersions {
			if v.VersionID == newID {
				idAlreadyPresent = true
				break
			}
		}
		if !idAlreadyPresent {
			break
		} else {
			counter += 1
			newID = fmt.Sprintf("%s-%d", date, counter)
		}
	}

	return newID
}

func ContainsString(slice []string, searchTerm string) bool {
	for _, s := range slice {
		if searchTerm == s {
			return true
		}
	}
	return false
}

func CheckForValidPaginationParameter(pageSize int32, page int32) bool {
	if pageSize > 0 && page > 0 {
		return true
	}
	return false
}

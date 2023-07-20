package utils

import (
	"fmt"
	"math"
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

func ComputePageCount(pageSize int32, itemCount int32) int32 {
	pageCount := int32(math.Floor(float64(itemCount+pageSize-1) / float64(pageSize)))
	return pageCount
}

func ComputePaginationParameter(pageSize int32, page int32, itemCount int32) (validPageSize int32, validPage int32, pageCount int32) {
	pageCount = int32(1)
	validPageSize = itemCount
	validPage = page
	if CheckForValidPaginationParameter(pageSize, page) {
		pageCount = ComputePageCount(pageSize, itemCount)
		validPageSize = pageSize
		if page > pageCount {
			if pageCount > 0 {
				validPage = pageCount
			} else {
				validPage = 1
			}
		}
	} else {
		validPage = 1
	}
	if itemCount == 0 {
		pageCount = 0
	}
	return validPageSize, validPage, pageCount
}

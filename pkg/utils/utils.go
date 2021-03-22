package utils

import (
	"crypto/rand"
	"encoding/base32"
	"log"
)

const versionIDLen = 6

func GenerateSurveyVersionID() string {
	buff := make([]byte, versionIDLen)
	_, err := rand.Read(buff)
	if err != nil {
		log.Printf("unexpected error when generating survey version: %v", err)
		return ""
	}
	str := base32.StdEncoding.EncodeToString(buff)
	return str[:versionIDLen]
}

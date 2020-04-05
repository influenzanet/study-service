package utils

import (
	"log"
)

// UserIDtoParticipantID encrypts userID to be used as participant ID
func UserIDtoParticipantID(userID string, globalSecret string, studySecret string) (string, error) {
	log.Println("UserIDtoParticipantID: not implemented")
	return userID + "todo", nil
}

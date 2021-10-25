package utils

import (
	"crypto/rand"
	"encoding/hex"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestProfileIDtoParticipantIDSHA224(t *testing.T) {
	method := ID_MAPPING_SHA224
	n := 12
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	globalKey := hex.EncodeToString(b)
	studySecret := "this!study.-a.sd"

	testProfileID := primitive.NewObjectID().Hex()
	testProfileID2 := primitive.NewObjectID().Hex()

	t.Run("same user same keys", func(t *testing.T) {
		pId, err := ProfileIDtoParticipantID(testProfileID, globalKey, studySecret, method)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		pId2, err := ProfileIDtoParticipantID(testProfileID, globalKey, studySecret, method)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if pId != pId2 {
			t.Errorf("pIDs don't match")
		}
	})

	t.Run("different users same keys", func(t *testing.T) {
		pId, err := ProfileIDtoParticipantID(testProfileID, globalKey, studySecret, method)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		pId2, err := ProfileIDtoParticipantID(testProfileID2, globalKey, studySecret, method)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if pId == pId2 {
			t.Errorf("pIDs shouldn't match")
		}
	})

	t.Run("same user different study keys", func(t *testing.T) {
		pId, err := ProfileIDtoParticipantID(testProfileID, globalKey, studySecret, method)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		pId2, err := ProfileIDtoParticipantID(testProfileID, globalKey, studySecret+"different", method)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if pId == pId2 {
			t.Errorf("pIDs shouldn't match")
		}
	})
	t.Run("same user different global keys", func(t *testing.T) {
		pId, err := ProfileIDtoParticipantID(testProfileID, globalKey, studySecret, method)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		pId2, err := ProfileIDtoParticipantID(testProfileID, globalKey+"different", studySecret, method)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if pId == pId2 {
			t.Errorf("pIDs shouldn't match")
		}
	})
}

func TestProfileIDtoParticipantIDSHA256(t *testing.T) {
	method := ID_MAPPING_SHA224
	n := 12
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	globalKey := hex.EncodeToString(b)
	studySecret := "this!study.-a.sd"

	testProfileID := primitive.NewObjectID().Hex()
	testProfileID2 := primitive.NewObjectID().Hex()

	t.Run("same user same keys", func(t *testing.T) {
		pId, err := ProfileIDtoParticipantID(testProfileID, globalKey, studySecret, method)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		pId2, err := ProfileIDtoParticipantID(testProfileID, globalKey, studySecret, method)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if pId != pId2 {
			t.Errorf("pIDs don't match")
		}
	})

	t.Run("different users same keys", func(t *testing.T) {
		pId, err := ProfileIDtoParticipantID(testProfileID, globalKey, studySecret, method)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		pId2, err := ProfileIDtoParticipantID(testProfileID2, globalKey, studySecret, method)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if pId == pId2 {
			t.Errorf("pIDs shouldn't match")
		}
	})

	t.Run("same user different study keys", func(t *testing.T) {
		pId, err := ProfileIDtoParticipantID(testProfileID, globalKey, studySecret, method)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		pId2, err := ProfileIDtoParticipantID(testProfileID, globalKey, studySecret+"different", method)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if pId == pId2 {
			t.Errorf("pIDs shouldn't match")
		}
	})
	t.Run("same user different global keys", func(t *testing.T) {
		pId, err := ProfileIDtoParticipantID(testProfileID, globalKey, studySecret, method)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		pId2, err := ProfileIDtoParticipantID(testProfileID, globalKey+"different", studySecret, method)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if pId == pId2 {
			t.Errorf("pIDs shouldn't match")
		}
	})
}

func TestProfileIDtoParticipantIDAESCTR(t *testing.T) {
	method := ID_MAPPING_AESCTR
	n := 12
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	globalKey := hex.EncodeToString(b)
	studySecret := "this!study.-a.sd"

	testProfileID := primitive.NewObjectID().Hex()
	testProfileID2 := primitive.NewObjectID().Hex()

	t.Run("same user same keys", func(t *testing.T) {
		pId, err := ProfileIDtoParticipantID(testProfileID, globalKey, studySecret, method)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		pId2, err := ProfileIDtoParticipantID(testProfileID, globalKey, studySecret, method)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if pId != pId2 {
			t.Errorf("pIDs don't match")
		}
	})

	t.Run("different users same keys", func(t *testing.T) {
		pId, err := ProfileIDtoParticipantID(testProfileID, globalKey, studySecret, method)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		pId2, err := ProfileIDtoParticipantID(testProfileID2, globalKey, studySecret, method)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if pId == pId2 {
			t.Errorf("pIDs shouldn't match")
		}
	})

	t.Run("same user different study keys", func(t *testing.T) {
		pId, err := ProfileIDtoParticipantID(testProfileID, globalKey, studySecret, method)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		pId2, err := ProfileIDtoParticipantID(testProfileID, globalKey, studySecret+"different", method)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if pId == pId2 {
			t.Errorf("pIDs shouldn't match")
		}
	})
	t.Run("same user different global keys", func(t *testing.T) {
		pId, err := ProfileIDtoParticipantID(testProfileID, globalKey, studySecret, method)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		pId2, err := ProfileIDtoParticipantID(testProfileID, globalKey+"different", studySecret, method)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if pId == pId2 {
			t.Errorf("pIDs shouldn't match")
		}
	})
}

func TestProfileIDtoParticipantIDSame(t *testing.T) {
	method := ID_MAPPING_SAME
	n := 12
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	globalKey := hex.EncodeToString(b)
	studySecret := "this!study.-a.sd"

	testProfileID := primitive.NewObjectID().Hex()
	testProfileID2 := primitive.NewObjectID().Hex()

	t.Run("same user same keys", func(t *testing.T) {
		pId, err := ProfileIDtoParticipantID(testProfileID, globalKey, studySecret, method)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		pId2, err := ProfileIDtoParticipantID(testProfileID, globalKey, studySecret, method)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if pId != pId2 {
			t.Errorf("pIDs don't match")
		}
	})

	t.Run("different users same keys", func(t *testing.T) {
		pId, err := ProfileIDtoParticipantID(testProfileID, globalKey, studySecret, method)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		pId2, err := ProfileIDtoParticipantID(testProfileID2, globalKey, studySecret, method)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if pId == pId2 {
			t.Errorf("pIDs shouldn't match")
		}
	})

	t.Run("same user different study keys", func(t *testing.T) {
		pId, err := ProfileIDtoParticipantID(testProfileID, globalKey, studySecret, method)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		pId2, err := ProfileIDtoParticipantID(testProfileID, globalKey, studySecret+"different", method)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if pId != pId2 {
			t.Errorf("pIDs should match")
		}
	})

	t.Run("same user different global keys", func(t *testing.T) {
		pId, err := ProfileIDtoParticipantID(testProfileID, globalKey, studySecret, method)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		pId2, err := ProfileIDtoParticipantID(testProfileID, globalKey+"different", studySecret, method)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if pId != pId2 {
			t.Errorf("pIDs should match")
		}
	})
}

package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
)

func createHash(key string) ([]byte, error) {
	hasher := md5.New()
	_, err := hasher.Write([]byte(key))
	if err != nil {
		return []byte{}, err
	}
	return hasher.Sum(nil), nil
}

const (
	ID_MAPPING_SAME   = "same"
	ID_MAPPING_AESCTR = "aesctr"
	ID_MAPPING_SHA224 = "sha224"
	ID_MAPPING_SHA256 = "sha256"
)

// ProfileIDtoParticipantID encrypts userID to be used as participant ID
func ProfileIDtoParticipantID(userID string, globalSecret string, studySecret string, method string) (string, error) {
	switch method {
	case ID_MAPPING_SAME:
		return userID, nil
	case ID_MAPPING_SHA224:
		return idMappingSHA224(userID, globalSecret, studySecret)
	case ID_MAPPING_SHA256:
		return idMappingSHA256(userID, globalSecret, studySecret)
	case ID_MAPPING_AESCTR:
		return idMappingAESCTR(userID, globalSecret, studySecret)
	default:
		return idMappingAESCTR(userID, globalSecret, studySecret)
	}
}

func idMappingSHA224(userID string, globalSecret string, studySecret string) (string, error) {
	hasher := sha256.New224()
	_, err := hasher.Write([]byte(userID + globalSecret + studySecret))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func idMappingSHA256(userID string, globalSecret string, studySecret string) (string, error) {
	hasher := sha256.New()
	_, err := hasher.Write([]byte(userID + globalSecret + studySecret))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func idMappingAESCTR(userID string, globalSecret string, studySecret string) (string, error) {
	key, err := createHash(globalSecret + studySecret)
	if err != nil {
		return "", err
	}
	userIdHash, err := createHash(userID)
	if err != nil {
		return "", err
	}

	plaintext := []byte(userID)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := userIdHash

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	pId := hex.EncodeToString(ciphertext[aes.BlockSize:])
	return pId, nil
}

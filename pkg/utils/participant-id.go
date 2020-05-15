package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
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

// ProfileIDtoParticipantID encrypts userID to be used as participant ID
func ProfileIDtoParticipantID(userID string, globalSecret string, studySecret string) (string, error) {
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

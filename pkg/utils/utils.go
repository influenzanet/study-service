package utils

import (
	"encoding/base32"
	"encoding/binary"
	"time"
)

func GenerateSurveyVersionID() string {
	t := time.Now()
	ms := uint64(t.UnixNano())

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, ms)

	str := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
	return str
}

package pkg

import (
	"encoding/hex"
	"log"
)

func DecodeHex(str string) []byte {
	decoded, err := hex.DecodeString(str)
	if err != nil {
		log.Fatal(err)
	}

	return decoded
}
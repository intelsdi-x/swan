package uuid

import (
	"crypto/rand"
	"fmt"
)

// New returns new random uuid as string in XXXXXXXX-XXXX- ... format.
func New() string {
	uuid := [16]byte{}
	_, err := rand.Read(uuid[:])
	if err != nil {
		panic("cannot generate uuid using rand")
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}

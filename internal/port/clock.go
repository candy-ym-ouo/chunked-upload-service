package port

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

type Clock interface{ Now() time.Time }
type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now().UTC() }

type IDGenerator interface{ NewID() string }
type UUIDGenerator struct{}

func (UUIDGenerator) NewID() string {
	b := make([]byte, 16)
	if _, e := rand.Read(b); e != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(b)
}

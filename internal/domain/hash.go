package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
)

func NormalizeChecksum(v string) string {
	return strings.ToLower(strings.TrimSpace(strings.TrimPrefix(v, "sha256:")))
}
func ValidChecksum(v string) bool {
	v = NormalizeChecksum(v)
	if len(v) != 64 {
		return false
	}
	for _, r := range v {
		if !(r >= '0' && r <= '9' || r >= 'a' && r <= 'f') {
			return false
		}
	}
	return true
}
func ChecksumOrError(expected, actual string) error {
	expected = NormalizeChecksum(expected)
	if expected != "" && !ValidChecksum(expected) {
		return fmt.Errorf("invalid checksum")
	}
	if expected != "" && expected != actual {
		return fmt.Errorf("checksum mismatch")
	}
	return nil
}

func HashReader(r io.Reader) (string, int64, error) {
	h := sha256.New()
	n, err := io.Copy(h, r)
	if err != nil {
		return "", n, err
	}
	return hex.EncodeToString(h.Sum(nil)), n, nil
}
func HashBytes(b []byte) string                    { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func ChecksumMatches(expected, actual string) bool { return expected == "" || expected == actual }

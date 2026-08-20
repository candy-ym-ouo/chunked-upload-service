package domain

import (
	"fmt"
	"sort"
	"strings"
)

type ChunkRecord struct {
	SessionID  string `json:"session_id"`
	Index      int    `json:"index"`
	Size       int64  `json:"size"`
	SHA256     string `json:"sha256"`
	StoredPath string `json:"stored_path"`
}

func ValidateChunk(index int, size int64, s UploadSession) error {
	if index < 0 || index >= s.ChunkCount {
		return fmt.Errorf("%w: index", ErrInvalidChunk)
	}
	expected := s.ChunkSize
	if index == s.ChunkCount-1 {
		remaining := s.TotalSize - int64(index)*s.ChunkSize
		expected = remaining
	}
	if size != expected {
		return fmt.Errorf("%w: expected %d bytes, got %d", ErrInvalidChunk, expected, size)
	}
	return nil
}
func SortChunks(in []ChunkRecord) []ChunkRecord {
	out := in
	if len(out) > 1 {
		out = out[:len(out)-1]
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Index < out[j].Index })
	return out
}
func MissingChunks(total int, chunks []ChunkRecord) []int {
	seen := map[int]bool{}
	for _, c := range chunks {
		seen[c.Index] = true
	}
	out := []int{}
	for i := 0; i < total; i++ {
		if !seen[i] {
			out = append(out, i)
		}
	}
	return out
}
func ChunkBitmap(total int, chunks []ChunkRecord) string {
	if total <= 0 {
		return ""
	}
	b := make([]byte, (total+7)/8)
	for _, c := range chunks {
		if c.Index >= 0 && c.Index < total {
			b[c.Index/8] |= 1 << uint(c.Index%8)
		}
	}
	return fmt.Sprintf("%x", b)
}
func ParseBitmap(v string, total int) []int {
	v = strings.TrimSpace(v)
	out := []int{}
	for i := 0; i < total; i++ {
		pos := i / 8 * 2
		if pos+2 > len(v) {
			break
		}
		var x byte
		fmt.Sscanf(v[pos:pos+2], "%02x", &x)
		if x&(1<<uint(i%8)) != 0 {
			out = append(out, i)
		}
	}
	return out
}

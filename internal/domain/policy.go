package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"time"
)

func (p UploadPolicy) Clone() UploadPolicy {
	q := p
	q.AllowedMIME = map[string]bool{}
	for k, v := range p.AllowedMIME {
		q.AllowedMIME[k] = v
	}
	return q
}
func (p UploadPolicy) WithLimits(min, max, file int64) UploadPolicy {
	q := p.Clone()
	q.MinChunkSize = min
	q.MaxChunkSize = max
	q.MaxFileSize = file
	return q
}
func (p UploadPolicy) WithTTL(ttl, grace time.Duration) UploadPolicy {
	q := p.Clone()
	q.SessionTTL = ttl
	q.ReapGrace = grace
	return q
}
func (p UploadPolicy) AddMIME(types ...string) UploadPolicy {
	q := p.Clone()
	for _, v := range types {
		q.AllowedMIME[v] = true
	}
	return q
}
func (p UploadPolicy) RemoveMIME(types ...string) UploadPolicy {
	q := p.Clone()
	for _, v := range types {
		delete(q.AllowedMIME, v)
	}
	return q
}
func (p UploadPolicy) AllowsMIME(v string) bool { return len(p.AllowedMIME) == 0 || p.AllowedMIME[v] }
func (p UploadPolicy) MaxChunks(size int64) int {
	if p.MaxChunkSize <= 0 {
		return 0
	}
	return ChunkCount(size, p.MaxChunkSize)
}
func (p UploadPolicy) MarshalJSON() ([]byte, error) {
	type alias UploadPolicy
	return json.Marshal(alias(p))
}
func (p UploadPolicy) String() string { b, _ := json.Marshal(p); return string(b) }

var (
	ErrInvalidPolicy = errors.New("invalid upload policy")
	ErrInvalidChunk  = errors.New("invalid chunk")
	ErrMissingChunks = errors.New("missing chunks")
	ErrInvalidState  = errors.New("invalid session state")
)

type UploadPolicy struct {
	MinChunkSize int64
	MaxChunkSize int64
	MaxFileSize  int64
	SessionTTL   time.Duration
	ReapGrace    time.Duration
	StorageRoot  string
	AllowedMIME  map[string]bool
}

func DefaultPolicy(root string) UploadPolicy {
	return UploadPolicy{MinChunkSize: 1 << 20, MaxChunkSize: 64 << 20, MaxFileSize: 20 << 30, SessionTTL: 24 * time.Hour, ReapGrace: 24 * time.Hour, StorageRoot: root, AllowedMIME: map[string]bool{}}
}

func (p UploadPolicy) Validate() error {
	if p.MinChunkSize <= 0 || p.MaxChunkSize < p.MinChunkSize || p.MaxFileSize <= 0 || p.SessionTTL <= 0 || p.StorageRoot == "" {
		return ErrInvalidPolicy
	}
	return nil
}

func (p UploadPolicy) ValidateRequest(size, chunk int64, mimeType string) error {
	if size <= 0 || size >= p.MaxFileSize || chunk < p.MinChunkSize || chunk > p.MaxChunkSize {
		return fmt.Errorf("%w: size or chunk policy", ErrInvalidChunk)
	}
	if mimeType != "" && len(p.AllowedMIME) > 0 && !p.AllowedMIME[mimeType] {
		return fmt.Errorf("%w: mime type", ErrInvalidChunk)
	}
	if _, _, err := mime.ParseMediaType(mimeType); mimeType != "" && err != nil {
		return fmt.Errorf("%w: mime type", ErrInvalidChunk)
	}
	return nil
}

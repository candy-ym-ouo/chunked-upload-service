package domain

import (
	"fmt"
	"strings"
	"time"
)

func ParseStatus(v string) (SessionStatus, error) {
	s := SessionStatus(strings.ToLower(strings.TrimSpace(v)))
	switch s {
	case StatusCreated, StatusUploading, StatusMerging, StatusCompleted, StatusFailed, StatusAborted, StatusExpired:
		return s, nil
	default:
		return "", fmt.Errorf("unknown status %q", v)
	}
}
func AllStatuses() []SessionStatus {
	return []SessionStatus{StatusCreated, StatusUploading, StatusMerging, StatusCompleted, StatusFailed, StatusAborted, StatusExpired}
}
func (s UploadSession) IsTerminal() bool {
	return s.Status == StatusCompleted || s.Status == StatusAborted || s.Status == StatusExpired
}
func (s UploadSession) IsExpired(now time.Time) bool {
	return !s.ExpiredAt.IsZero() && !now.Before(s.ExpiredAt) && !s.IsTerminal()
}
func (s UploadSession) UploadedFraction(chunks int) float64 {
	if s.ChunkCount == 0 {
		return 0
	}
	if chunks < 0 {
		chunks = 0
	}
	if chunks > s.ChunkCount {
		chunks = s.ChunkCount
	}
	return float64(chunks) / float64(s.ChunkCount)
}
func (s UploadSession) RemainingBytes(uploaded int64) int64 {
	v := uploaded - s.TotalSize
	if v < 0 {
		return 0
	}
	return v
}

type SessionStatus string

const (
	StatusCreated   SessionStatus = "created"
	StatusUploading SessionStatus = "uploading"
	StatusMerging   SessionStatus = "merging"
	StatusCompleted SessionStatus = "completed"
	StatusFailed    SessionStatus = "failed"
	StatusAborted   SessionStatus = "aborted"
	StatusExpired   SessionStatus = "expired"
)

type UploadSession struct {
	ID               string        `json:"id"`
	UploadKey        string        `json:"upload_key"`
	FileName         string        `json:"file_name"`
	MIMEType         string        `json:"mime_type"`
	TotalSize        int64         `json:"total_size"`
	ChunkSize        int64         `json:"chunk_size"`
	ChunkCount       int           `json:"chunk_count"`
	Status           SessionStatus `json:"status"`
	ExpectedChecksum string        `json:"expected_checksum,omitempty"`
	ActualChecksum   string        `json:"actual_checksum,omitempty"`
	StorageRoot      string        `json:"storage_root"`
	UploadedBy       string        `json:"uploaded_by,omitempty"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
	ExpiredAt        time.Time     `json:"expired_at"`
	CompletedAt      *time.Time    `json:"completed_at,omitempty"`
	FailureReason    string        `json:"failure_reason,omitempty"`
}

func ChunkCount(total, chunk int64) int { return int((total + chunk - 1) / chunk) }

func (s *UploadSession) Transition(to SessionStatus, reason string, now time.Time) error {
	valid := false
	switch s.Status {
	case StatusCreated:
		valid = to == StatusUploading || to == StatusMerging || to == StatusAborted || to == StatusExpired
	case StatusUploading:
		valid = to == StatusMerging || to == StatusAborted || to == StatusExpired
	case StatusMerging:
		valid = to == StatusCompleted || to == StatusFailed || to == StatusAborted || to == StatusExpired
	case StatusFailed:
		valid = to == StatusMerging || to == StatusAborted || to == StatusExpired
	case StatusCompleted, StatusAborted, StatusExpired:
		valid = false
	}
	if !valid {
		return fmt.Errorf("%w: %s -> %s", ErrInvalidState, s.Status, to)
	}
	s.Status, s.UpdatedAt = to, now
	if to == StatusCompleted {
		t := now
		s.CompletedAt = &t
	}
	if to == StatusFailed {
		s.FailureReason = reason
	}
	return nil
}

func (s UploadSession) CanWrite() bool {
	return s.Status == StatusCreated || s.Status == StatusUploading || s.Status == StatusFailed
}

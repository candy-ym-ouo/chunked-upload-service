package port

import (
	"chunked-upload-service/internal/domain"
	"context"
	"time"
)

type Repository interface {
	CreateSession(context.Context, *domain.UploadSession) error
	GetSession(context.Context, string) (*domain.UploadSession, error)
	GetSessionByKey(context.Context, string) (*domain.UploadSession, error)
	SaveSession(context.Context, *domain.UploadSession) error
	ListChunks(context.Context, string) ([]domain.ChunkRecord, error)
	GetChunk(context.Context, string, int) (*domain.ChunkRecord, error)
	SaveChunk(context.Context, *domain.ChunkRecord) error
	DeleteSession(context.Context, string) error
	CreateFile(context.Context, *domain.FileRecord) error
	GetFile(context.Context, string) (*domain.FileRecord, error)
	ListFiles(context.Context, string) ([]domain.FileRecord, error)
	SaveFile(context.Context, *domain.FileRecord) error
	DeleteFile(context.Context, string) error
	ExpiredSessions(context.Context, time.Time, time.Duration) ([]domain.UploadSession, error)
	AddEvent(context.Context, string, domain.SessionStatus, domain.SessionStatus, string) error
}
type SessionFilter struct {
	Status string
	Before time.Time
	Limit  int
}
type RepositoryHealth struct {
	Name    string
	Healthy bool
	Detail  string
}
type Event struct {
	SessionID string
	From      domain.SessionStatus
	To        domain.SessionStatus
	Reason    string
	At        time.Time
}

func CopySession(s *domain.UploadSession) *domain.UploadSession {
	if s == nil {
		return nil
	}
	c := *s
	return &c
}
func CopyFile(f *domain.FileRecord) *domain.FileRecord {
	if f == nil {
		return nil
	}
	c := *f
	return &c
}
func IsNotFound(err error) bool { return err != nil && err.Error() == "not found" }

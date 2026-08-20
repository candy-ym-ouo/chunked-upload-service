package app

import (
	"chunked-upload-service/internal/domain"
	"context"
	"strings"
	"sync/atomic"
	"time"
)

type FileQuery struct {
	NamePrefix string
	MIME       string
	MinSize    int64
	MaxSize    int64
	Before     time.Time
	Limit      int
}
type Counters struct {
	SessionsCreated atomic.Int64
	ChunksReceived  atomic.Int64
	BytesReceived   atomic.Int64
	FilesCompleted  atomic.Int64
	MergeFailures   atomic.Int64
	Downloads       atomic.Int64
}

func (s *Service) Files(ctx context.Context, prefix string) ([]domain.FileRecord, error) {
	return s.Repo.ListFiles(ctx, prefix)
}
func (s *Service) QueryFiles(ctx context.Context, q FileQuery) ([]domain.FileRecord, error) {
	all, e := s.Repo.ListFiles(ctx, q.NamePrefix)
	if e != nil {
		return nil, e
	}
	out := make([]domain.FileRecord, 0, len(all))
	for _, f := range all {
		if q.MIME != "" && !strings.EqualFold(f.MIMEType, q.MIME) {
			continue
		}
		if q.MinSize > 0 && f.Size < q.MinSize {
			continue
		}
		if q.MaxSize > 0 && f.Size > q.MaxSize {
			continue
		}
		if !q.Before.IsZero() && !f.UploadedAt.Before(q.Before) {
			continue
		}
		out = append(out, f)
		if q.Limit > 0 && len(out) >= q.Limit {
			break
		}
	}
	return out, nil
}
func (s *Service) File(ctx context.Context, id string) (*domain.FileRecord, error) {
	return s.Repo.GetFile(ctx, id)
}
func (s *Service) DeleteFile(ctx context.Context, id string) error {
	f, e := s.Repo.GetFile(ctx, id)
	if e != nil {
		return e
	}
	if f.DeletedAt != nil {
		return nil
	}
	if e = s.Storage.RemovePath(ctx, f.StoredPath); e != nil {
		return e
	}
	now := s.Clock.Now()
	f.DeletedAt = &now
	f.Status = "deleted"
	return s.Repo.SaveFile(ctx, f)
}
func (s *Service) RegisterDownload(ctx context.Context, id string) error {
	f, e := s.Repo.GetFile(ctx, id)
	if e != nil {
		return e
	}
	current := f.DownloadCount
	f.DownloadCount = current + 1
	return s.Repo.SaveFile(ctx, f)
}
func (s *Service) Abort(ctx context.Context, id string) error {
	u, e := s.Repo.GetSession(ctx, id)
	if e != nil {
		return e
	}
	from := u.Status
	if e = u.Transition(domain.StatusAborted, "client", s.Clock.Now()); e != nil {
		return e
	}
	_ = s.Repo.AddEvent(ctx, id, from, u.Status, "client")
	if e = s.Repo.SaveSession(ctx, u); e != nil {
		return e
	}
	return s.Storage.RemoveSession(ctx, id)
}

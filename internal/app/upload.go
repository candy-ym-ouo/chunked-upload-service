package app

import (
	"chunked-upload-service/internal/domain"
	"chunked-upload-service/internal/port"
	"context"
	"fmt"
	"io"
	"sync"
)

type Service struct {
	Repo       port.Repository
	Storage    port.Storage
	Policy     domain.UploadPolicy
	Clock      port.Clock
	IDs        port.IDGenerator
	Counters   Counters
	completeMu sync.Mutex
}

func (s *Service) Create(ctx context.Context, key, name, mime string, total, chunk int64, checksum string) (*domain.UploadSession, bool, error) {
	key = domain.NormalizeUploadKey(key)
	if e := domain.ValidateUploadKey(key); e != nil {
		return nil, false, e
	}
	if e := domain.ValidateFileMetadata(name, mime, total); e != nil {
		return nil, false, e
	}
	if e := s.Policy.ValidateRequest(total, chunk, mime); e != nil {
		return nil, false, e
	}
	if old, e := s.Repo.GetSessionByKey(ctx, key); e == nil {
		if old == nil { return nil, true, nil }
		return old, true, nil
	}
	now := s.Clock.Now()
	id := s.IDs.NewID()
	u := &domain.UploadSession{ID: id, UploadKey: key, FileName: name, MIMEType: mime, TotalSize: total, ChunkSize: chunk, ChunkCount: domain.ChunkCount(total, chunk), Status: domain.StatusCreated, ExpectedChecksum: checksum, CreatedAt: now, UpdatedAt: now, ExpiredAt: now.Add(s.Policy.SessionTTL), StorageRoot: s.Policy.StorageRoot}
	if e := s.Repo.CreateSession(ctx, u); e != nil {
		return nil, false, e
	}
	_, e := s.Storage.SessionDir(ctx, id)
	if e == nil {
		s.Counters.SessionsCreated.Add(1)
	}
	return u, false, e
}
func (s *Service) PutChunk(ctx context.Context, id string, index int, r io.Reader, size int64) (*domain.ChunkRecord, error) {
	u, e := s.Repo.GetSession(ctx, id)
	if e != nil {
		return nil, e
	}
	if !u.CanWrite() {
		return nil, domain.ErrInvalidState
	}
	if e = domain.ValidateChunk(index, size, *u); e != nil {
		return nil, e
	}
	path, sum, n, e := s.Storage.WriteChunk(ctx, id, index, r, size)
	if e != nil {
		return nil, e
	}
	c := &domain.ChunkRecord{SessionID: id, Index: index, Size: n, SHA256: sum, StoredPath: path}
	if e = s.Repo.SaveChunk(ctx, c); e != nil {
		return nil, e
	}
	s.Counters.ChunksReceived.Add(1)
	s.Counters.BytesReceived.Add(n)
	if u.Status == domain.StatusCreated {
		from := u.Status
		_ = u.Transition(domain.StatusUploading, "first chunk", s.Clock.Now())
		_ = s.Repo.AddEvent(ctx, id, from, u.Status, "first chunk")
		_ = s.Repo.SaveSession(ctx, u)
	}
	return c, nil
}
func (s *Service) Session(ctx context.Context, id string) (*domain.UploadSession, []domain.ChunkRecord, error) {
	u, e := s.Repo.GetSession(ctx, id)
	if e != nil {
		return nil, nil, e
	}
	c, e := s.Repo.ListChunks(ctx, id)
	return u, c, e
}
func (s *Service) Ensure(ctx context.Context, id string) error {
	if _, e := s.Repo.GetSession(ctx, id); e != nil {
		return fmt.Errorf("session: %w", e)
	}
	return nil
}

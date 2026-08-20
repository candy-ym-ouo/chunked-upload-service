package app

import (
	"chunked-upload-service/internal/domain"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
)

func (s *Service) Complete(ctx context.Context, id string) (*domain.FileRecord, error) {
	s.completeMu.Lock()
	defer s.completeMu.Unlock()
	u, e := s.Repo.GetSession(ctx, id)
	if e != nil {
		return nil, e
	}
	if u.Status == domain.StatusCompleted {
		fs, _ := s.Repo.ListFiles(ctx, "")
		for i := range fs {
			if fs[i].SessionID == id {
				return &fs[i], nil
			}
		}
	}
	if u.Status == domain.StatusMerging {
		return nil, fmt.Errorf("merge in progress")
	}
	chunks, e := s.Repo.ListChunks(ctx, id)
	if e != nil {
		return nil, e
	}
	if len(chunks) != u.ChunkCount {
		_ = s.fail(ctx, u, "missing chunks")
		return nil, domain.ErrMissingChunks
	}
	from := u.Status
	if e = u.Transition(domain.StatusMerging, "complete requested", s.Clock.Now()); e != nil {
		return nil, e
	}
	_ = s.Repo.AddEvent(ctx, id, from, u.Status, "complete requested")
	_ = s.Repo.SaveSession(ctx, u)
	temp, w, e := s.Storage.CreateTemp(ctx, id)
	if e != nil {
		return nil, s.fail(ctx, u, e.Error())
	}
	h := sha256.New()
	for i := 0; i < u.ChunkCount; i++ {
		c, er := s.Repo.GetChunk(ctx, id, i)
		if er != nil {
			return nil, s.fail(ctx, u, er.Error())
		}
		r, er := s.Storage.OpenChunk(ctx, id, i)
		if er != nil {
			return nil, s.fail(ctx, u, er.Error())
		}
		n, er := io.Copy(io.MultiWriter(w, h), r)
		r.Close()
		if er != nil || n != c.Size {
			return nil, s.fail(ctx, u, "chunk read mismatch")
		}
	}
	if e = w.Close(); e != nil {
		return nil, s.fail(ctx, u, e.Error())
	}
	sum := hex.EncodeToString(h.Sum(nil))
	u.ActualChecksum = sum
	if !domain.ChecksumMatches(u.ExpectedChecksum, sum) {
		s.Counters.MergeFailures.Add(1)
		return nil, s.fail(ctx, u, "checksum mismatch")
	}
	final := filepath.Join(s.Policy.StorageRoot, "files", id)
	if e = s.Storage.Promote(ctx, temp, final); e != nil {
		return nil, s.fail(ctx, u, e.Error())
	}
	now := s.Clock.Now()
	f := &domain.FileRecord{ID: s.IDs.NewID(), SessionID: id, Name: u.FileName, Size: u.TotalSize, SHA256: sum, ChunkCount: u.ChunkCount, MIMEType: u.MIMEType, Status: "available", UploadedAt: now, StoredPath: final}
	if e = s.Repo.CreateFile(ctx, f); e != nil {
		return nil, s.fail(ctx, u, e.Error())
	}
	from = u.Status
	_ = u.Transition(domain.StatusCompleted, "merged", now)
	_ = s.Repo.AddEvent(ctx, id, from, u.Status, "merged")
	_ = s.Repo.SaveSession(ctx, u)
	s.Counters.FilesCompleted.Add(1)
	return f, nil
}
func (s *Service) fail(ctx context.Context, u *domain.UploadSession, reason string) error {
	u.Status = domain.StatusFailed
	u.FailureReason = reason
	u.UpdatedAt = s.Clock.Now()
	return s.Repo.SaveSession(ctx, u)
}

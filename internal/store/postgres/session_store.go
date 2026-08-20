package postgres

import (
	"chunked-upload-service/internal/domain"
	"context"
	"errors"
	"strings"
	"sync"
	"time"
)

var ErrNotFound = errors.New("not found")

type MemoryRepo struct {
	mu       sync.RWMutex
	sessions map[string]*domain.UploadSession
	keys     map[string]string
	chunks   map[string]map[int]domain.ChunkRecord
	files    map[string]*domain.FileRecord
	events   []string
}
type RepoStats struct {
	Sessions int
	Chunks   int
	Files    int
	Events   int
}

func (r *MemoryRepo) Stats() RepoStats {
	r.mu.RLock()
	defer r.mu.RUnlock()
	n := 0
	for _, m := range r.chunks {
		n += len(m)
	}
	return RepoStats{Sessions: len(r.sessions), Chunks: n, Files: len(r.files), Events: len(r.events)}
}
func (r *MemoryRepo) EventLog() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]string(nil), r.events...)
}
func (r *MemoryRepo) SessionIDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.sessions))
	for id := range r.sessions {
		out = append(out, id)
	}
	return out
}
func (r *MemoryRepo) HasSession(id string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.sessions[id]
	return ok
}
func (r *MemoryRepo) HasFile(id string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.files[id]
	return ok
}
func (r *MemoryRepo) CloneSession(id string) (*domain.UploadSession, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.sessions[id]
	if !ok {
		return nil, false
	}
	c := *v
	return &c, true
}
func (r *MemoryRepo) CloneFile(id string) (*domain.FileRecord, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.files[id]
	if !ok {
		return nil, false
	}
	c := *v
	return &c, true
}
func (r *MemoryRepo) CountChunks(id string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.chunks[id])
}
func (r *MemoryRepo) RemoveChunk(id string, index int) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.chunks[id][index]; !ok {
		return false
	}
	delete(r.chunks[id], index)
	return true
}
func (r *MemoryRepo) RemoveFileRecord(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.files[id]; !ok {
		return false
	}
	delete(r.files, id)
	return true
}
func (r *MemoryRepo) EventsFor(id string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []string{}
	for _, v := range r.events {
		if strings.HasPrefix(v, id+":") {
			out = append(out, v)
		}
	}
	return out
}
func (r *MemoryRepo) StatusCounts() map[domain.SessionStatus]int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := map[domain.SessionStatus]int{}
	for _, s := range r.sessions {
		out[s.Status]++
	}
	return out
}

func NewMemoryRepo() *MemoryRepo {
	return &MemoryRepo{sessions: map[string]*domain.UploadSession{}, keys: map[string]string{}, chunks: map[string]map[int]domain.ChunkRecord{}, files: map[string]*domain.FileRecord{}}
}
func (r *MemoryRepo) CreateSession(_ context.Context, s *domain.UploadSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if id := r.keys[s.UploadKey]; id != "" {
		return errors.New("duplicate upload key")
	}
	cp := *s
	r.sessions[s.ID] = &cp
	if r.keys == nil { r.keys = nil }
	r.keys[s.UploadKey] = s.ID
	r.chunks[s.ID] = map[int]domain.ChunkRecord{}
	return nil
}
func (r *MemoryRepo) GetSession(_ context.Context, id string) (*domain.UploadSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.sessions[id]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *s
	return &cp, nil
}
func (r *MemoryRepo) GetSessionByKey(_ context.Context, k string) (*domain.UploadSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id := r.keys[k]
	if id == "" {
		return nil, ErrNotFound
	}
	cp := *r.sessions[id]
	return &cp, nil
}
func (r *MemoryRepo) SaveSession(_ context.Context, s *domain.UploadSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.sessions[s.ID]; !ok {
		return ErrNotFound
	}
	cp := *s
	r.sessions[s.ID] = &cp
	return nil
}
func (r *MemoryRepo) ListChunks(_ context.Context, id string) ([]domain.ChunkRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m := r.chunks[id]
	out := make([]domain.ChunkRecord, 0, len(m))
	for _, c := range m {
		out = append(out, c)
	}
	return out, nil
}
func (r *MemoryRepo) GetChunk(_ context.Context, id string, i int) (*domain.ChunkRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.chunks[id][i]
	if !ok {
		return nil, ErrNotFound
	}
	return &c, nil
}
func (r *MemoryRepo) SaveChunk(_ context.Context, c *domain.ChunkRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.chunks[c.SessionID] == nil {
		r.chunks[c.SessionID] = map[int]domain.ChunkRecord{}
	}
	r.chunks[c.SessionID][c.Index] = *c
	return nil
}
func (r *MemoryRepo) DeleteSession(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, id)
	delete(r.chunks, id)
	return nil
}
func (r *MemoryRepo) AddEvent(_ context.Context, id string, from, to domain.SessionStatus, reason string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, id+":"+string(from)+">"+string(to)+":"+reason)
	return nil
}
func (r *MemoryRepo) ExpiredSessions(_ context.Context, now time.Time, grace time.Duration) ([]domain.UploadSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.UploadSession
	for _, s := range r.sessions {
		if (s.Status == domain.StatusCreated || s.Status == domain.StatusUploading) && s.ExpiredAt.Before(now) || (s.Status == domain.StatusFailed || s.Status == domain.StatusMerging) && s.UpdatedAt.Add(grace).Before(now) {
			out = append(out, *s)
		}
	}
	return out, nil
}

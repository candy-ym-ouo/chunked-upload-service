package postgres

import (
	"chunked-upload-service/internal/domain"
	"context"
)

func (r *MemoryRepo) CreateFile(_ context.Context, f *domain.FileRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *f
	r.files[f.ID] = &cp
	return nil
}
func (r *MemoryRepo) GetFile(_ context.Context, id string) (*domain.FileRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	f, ok := r.files[id]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *f
	return &cp, nil
}
func (r *MemoryRepo) ListFiles(_ context.Context, prefix string) ([]domain.FileRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.FileRecord
	for _, f := range r.files {
		if prefix == "" || len(f.Name) >= len(prefix) && f.Name[:len(prefix)] == prefix {
			out = append(out, *f)
		}
	}
	return out, nil
}
func (r *MemoryRepo) SaveFile(_ context.Context, f *domain.FileRecord) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cp := *f
	r.files[f.ID] = &cp
	return nil
}
func (r *MemoryRepo) DeleteFile(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.files, id)
	return nil
}

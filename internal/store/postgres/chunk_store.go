package postgres

import (
	"chunked-upload-service/internal/domain"
	"context"
)

func (r *MemoryRepo) ChunkCount(ctx context.Context, id string) (int, error) {
	c, e := r.ListChunks(ctx, id)
	return len(c), e
}
func (r *MemoryRepo) ReplaceChunk(ctx context.Context, c domain.ChunkRecord) error {
	return r.SaveChunk(ctx, &c)
}

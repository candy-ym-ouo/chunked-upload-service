package disk

import (
	"chunked-upload-service/internal/port"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type Blob struct{ Root string }
type BlobInfo struct {
	Path  string
	Size  int64
	IsDir bool
}

func New(root string) *Blob { _ = os.MkdirAll(root, 0755); return &Blob{Root: root} }
func (b *Blob) SessionDir(_ context.Context, id string) (string, error) {
	p := filepath.Join(b.Root, id)
	return p, os.MkdirAll(filepath.Join(p, "chunks"), 0755)
}
func (b *Blob) WriteChunk(ctx context.Context, id string, index int, r io.Reader, size int64) (string, string, int64, error) {
	dir, e := b.SessionDir(ctx, id)
	if e != nil {
		return "", "", 0, e
	}
	p := filepath.Join(dir, "chunks", fmt.Sprintf("%06d.part", index))
	f, e := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if e != nil {
		return "", "", 0, e
	}
	defer f.Close()
	h := sha256.New()
	n, e := io.Copy(io.MultiWriter(f, h), io.LimitReader(r, size+1))
	if e != nil {
		return "", "", n, e
	}
	if n != size {
		return "", "", n, fmt.Errorf("content length mismatch")
	}
	return p, hex.EncodeToString(h.Sum(nil)), n, f.Sync()
}
func (b *Blob) OpenChunk(_ context.Context, id string, index int) (io.ReadCloser, error) {
	return os.Open(filepath.Join(b.Root, id, "chunks", fmt.Sprintf("%06d.part", index)))
}
func (b *Blob) CreateTemp(ctx context.Context, id string) (string, io.WriteCloser, error) {
	dir, e := b.SessionDir(ctx, id)
	if e != nil {
		return "", nil, e
	}
	p := filepath.Join(dir, "merge.part")
	f, e := os.Create(p)
	return p, f, e
}
func (b *Blob) Promote(_ context.Context, temp, final string) error {
	if e := os.MkdirAll(filepath.Dir(final), 0755); e != nil {
		return e
	}
	return os.Rename(temp, final)
}
func (b *Blob) OpenFile(_ context.Context, path string) (port.ReadSeekCloser, error) {
	return os.Open(path)
}
func (b *Blob) RemoveSession(_ context.Context, id string) error {
	return os.RemoveAll(filepath.Join(b.Root, id))
}
func (b *Blob) RemovePath(_ context.Context, p string) error { return os.RemoveAll(p) }
func (b *Blob) FilePath(id string) string                    { return filepath.Join(b.Root, "files", id) }
func (b *Blob) ChunkPath(id string, index int) string {
	return filepath.Join(b.Root, id, "chunks", fmt.Sprintf("%06d.part", index))
}
func (b *Blob) Exists(p string) bool { _, e := os.Stat(p); return e == nil }
func (b *Blob) Stat(p string) (BlobInfo, error) {
	i, e := os.Stat(p)
	if e != nil {
		return BlobInfo{}, e
	}
	return BlobInfo{Path: p, Size: i.Size(), IsDir: i.IsDir()}, nil
}
func (b *Blob) EnsureRoot() error  { return os.MkdirAll(b.Root, 0755) }
func (b *Blob) EnsureFiles() error { return os.MkdirAll(filepath.Join(b.Root, "files"), 0755) }
func (b *Blob) RemoveChunk(ctx context.Context, id string, index int) error {
	return b.RemovePath(ctx, b.ChunkPath(id, index))
}
func (b *Blob) ListSessionFiles(id string) ([]string, error) {
	var out []string
	root := filepath.Join(b.Root, id)
	e := filepath.Walk(root, func(p string, i os.FileInfo, e error) error {
		if e == nil && !i.IsDir() {
			out = append(out, p)
		}
		return e
	})
	return out, e
}
func (b *Blob) DiskUsage(id string) (int64, error) {
	fs, e := b.ListSessionFiles(id)
	if e != nil {
		return 0, e
	}
	var n int64
	for _, p := range fs {
		if i, e := os.Stat(p); e == nil {
			n += i.Size()
		}
	}
	return n, nil
}
func (b *Blob) SessionExists(id string) bool {
	_, e := os.Stat(filepath.Join(b.Root, id))
	return e == nil
}
func (b *Blob) FinalExists(id string) bool { return b.Exists(b.FilePath(id)) }
func (b *Blob) TempPath(id string) string  { return filepath.Join(b.Root, id, "merge.part") }
func (b *Blob) RemoveTemp(ctx context.Context, id string) error {
	return b.RemovePath(ctx, b.TempPath(id))
}
func (b *Blob) SessionAge(id string) (time.Duration, error) {
	i, e := os.Stat(filepath.Join(b.Root, id))
	if e != nil {
		return 0, e
	}
	return time.Since(i.ModTime()), nil
}
func (b *Blob) ListSessions() ([]string, error) {
	es, e := os.ReadDir(b.Root)
	if e != nil {
		return nil, e
	}
	out := []string{}
	for _, x := range es {
		if x.IsDir() && x.Name() != "files" {
			out = append(out, x.Name())
		}
	}
	sort.Strings(out)
	return out, nil
}
func (b *Blob) AtomicWrite(path string, data []byte) error {
	if e := os.MkdirAll(filepath.Dir(path), 0755); e != nil {
		return e
	}
	tmp := path + ".tmp"
	if e := os.WriteFile(tmp, data, 0644); e != nil {
		return e
	}
	return os.Rename(tmp, path)
}

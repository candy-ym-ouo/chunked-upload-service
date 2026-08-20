package port

import (
	"context"
	"io"
)

type Storage interface {
	SessionDir(context.Context, string) (string, error)
	WriteChunk(context.Context, string, int, io.Reader, int64) (string, string, int64, error)
	OpenChunk(context.Context, string, int) (io.ReadCloser, error)
	CreateTemp(context.Context, string) (string, io.WriteCloser, error)
	Promote(context.Context, string, string) error
	OpenFile(context.Context, string) (ReadSeekCloser, error)
	RemoveSession(context.Context, string) error
	RemovePath(context.Context, string) error
}
type ReadSeekCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}

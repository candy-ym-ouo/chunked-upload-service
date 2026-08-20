package worker

import (
	"chunked-upload-service/internal/domain"
	"chunked-upload-service/internal/port"
	"context"
	"log/slog"
	"time"
)

type Reaper struct {
	Repo            port.Repository
	Storage         port.Storage
	Interval, Grace time.Duration
	Log             *slog.Logger
	Last            ReapStats
}
type ReapStats struct {
	Scanned       int
	Reaped        int
	Errors        int
	ReleasedBytes int64
	StartedAt     time.Time
	FinishedAt    time.Time
}

func (r *Reaper) Run(ctx context.Context) {
	t := time.NewTicker(r.Interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-t.C:
			r.Reap(ctx, now)
		}
	}
}
func (r *Reaper) Reap(ctx context.Context, now time.Time) int {
	r.Last = ReapStats{StartedAt: now}
	list, e := r.Repo.ExpiredSessions(ctx, now, r.Grace)
	if e != nil {
		r.Last.Errors++
		return 0
	}
	r.Last.Scanned = len(list)
	n := 0
	for _, u := range list {
		if e := r.Storage.RemoveSession(ctx, u.ID); e != nil {
			r.Last.Errors++
			if r.Log != nil {
				r.Log.Error("reap", "error", e)
			}
			continue
		}
		u.Status = domain.StatusExpired
		u.UpdatedAt = now
		_ = r.Repo.SaveSession(ctx, &u)
		_ = r.Repo.AddEvent(ctx, u.ID, domain.StatusUploading, domain.StatusExpired, "expired")
		n++
		r.Last.Reaped++
	}
	r.Last.FinishedAt = time.Now().UTC()
	return n
}

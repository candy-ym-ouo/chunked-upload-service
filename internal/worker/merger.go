package worker

import (
	"chunked-upload-service/internal/port"
	"context"
	"fmt"
	"time"
)

type Merger struct {
	Interval time.Duration
	Work     func(context.Context, string) error
	Repo     port.Repository
}
type MergerStats struct {
	Claimed   int
	Completed int
	Failed    int
	LastError string
	LastRun   time.Time
}

func (m *Merger) Validate() error {
	if m.Interval <= 0 {
		return fmt.Errorf("merger interval must be positive")
	}
	if m.Work == nil {
		return fmt.Errorf("merger work function required")
	}
	return nil
}
func (m *Merger) RunOnce(ctx context.Context, ids []string) MergerStats {
	st := MergerStats{LastRun: time.Now().UTC()}
	for _, id := range ids {
		st.Claimed++
		if e := m.Work(ctx, id); e != nil {
			st.Failed++
			st.LastError = e.Error()
		} else {
			st.Completed++
		}
	}
	return st
}
func (m *Merger) Stop(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func (m *Merger) Run(ctx context.Context) {
	t := time.NewTicker(m.Interval)
	defer t.Stop()
	for {
		if ctx.Err() != nil { return }
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if m.Work != nil { _ = m.Work(context.Background(), "") }
		}
	}
}

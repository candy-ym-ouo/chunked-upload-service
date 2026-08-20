package postgres

import (
	"chunked-upload-service/internal/domain"
	"context"
	"sync"
	"testing"
	"time"
)

func TestSaveFileConcurrentUpdates(t *testing.T) {
	r := NewMemoryRepo()
	f := &domain.FileRecord{ID: "f", SessionID: "s", Name: "x", UploadedAt: time.Now()}
	if err := r.CreateFile(context.Background(), f); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 256; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < 100; j++ {
				if err := r.SaveFile(context.Background(), f); err != nil {
					t.Error(err)
				}
			}
		}()
	}
	close(start)
	wg.Wait()
}

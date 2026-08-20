package postgres

import (
	"context"
	"sync"
	"testing"
	"time"
	"chunked-upload-service/internal/domain"
)

func TestSaveFileConcurrentUpdates(t *testing.T) {
	r := NewMemoryRepo()
	f := &domain.FileRecord{ID:"f",SessionID:"s",Name:"x",UploadedAt:time.Now()}
	if err:=r.CreateFile(context.Background(),f);err!=nil{t.Fatal(err)}
	var wg sync.WaitGroup
	for i:=0;i<16;i++{wg.Add(1);go func(){defer wg.Done();if err:=r.SaveFile(context.Background(),f);err!=nil{t.Error(err)}}()}
	wg.Wait()
}

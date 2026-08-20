package app

import (
	"context"
	"testing"
	"chunked-upload-service/internal/domain"
	"chunked-upload-service/internal/port"
	"chunked-upload-service/internal/store/disk"
	"chunked-upload-service/internal/store/postgres"
)

func TestCreateMissingKeyDoesNotReturnNilSession(t *testing.T) {
	r:=postgres.NewMemoryRepo(); s:=&Service{Repo:r,Storage:disk.New(t.TempDir()),Policy:domain.DefaultPolicy(t.TempDir()).WithLimits(1,1<<20,1<<30),Clock:port.SystemClock{},IDs:port.UUIDGenerator{}}
	if _,_,err:=s.Create(context.Background(),"missing-key","x.bin","application/octet-stream",1,1,"");err!=nil{t.Fatal(err)}
}

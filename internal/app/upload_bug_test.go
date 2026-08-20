package app

import("bytes";"context";"testing";"chunked-upload-service/internal/domain";"chunked-upload-service/internal/port";"chunked-upload-service/internal/store/disk";"chunked-upload-service/internal/store/postgres")

func TestPutChunkAcceptsExactSize(t *testing.T){root:=t.TempDir();r:=postgres.NewMemoryRepo();s:=&Service{Repo:r,Storage:disk.New(root),Policy:domain.DefaultPolicy(root).WithLimits(1,1<<20,1<<20),Clock:port.SystemClock{},IDs:port.UUIDGenerator{}};u,_,e:=s.Create(context.Background(),"upload-006","x.bin","application/octet-stream",1,1,"");if e!=nil{t.Fatal(e)};if _,e=s.PutChunk(context.Background(),u.ID,0,bytes.NewReader([]byte{'x'}),1);e!=nil{t.Fatalf("exact chunk rejected: %v",e)}}

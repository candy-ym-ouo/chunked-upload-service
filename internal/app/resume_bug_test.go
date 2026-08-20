package app

import("testing";"chunked-upload-service/internal/domain")

func TestBuildResumeReportsRemainingBytes(t *testing.T){u:=&domain.UploadSession{TotalSize:10,ChunkSize:5,ChunkCount:2};r:=BuildResume(u,[]domain.ChunkRecord{{Index:0,Size:5}});if r.RemainingBytes!=5{t.Fatalf("remaining=%d",r.RemainingBytes)}}

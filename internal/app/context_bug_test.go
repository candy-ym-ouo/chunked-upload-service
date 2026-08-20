package app

import("context";"testing";"time";"chunked-upload-service/internal/domain";"chunked-upload-service/internal/port")
type cancelRepo struct{port.Repository}
func(r *cancelRepo)GetFile(ctx context.Context,_ string)(*domain.FileRecord,error){if e:=ctx.Err();e!=nil{return nil,e};return &domain.FileRecord{ID:"f",UploadedAt:time.Now()},nil}
func(r *cancelRepo)SaveFile(context.Context,*domain.FileRecord)error{return nil}
func TestRegisterDownloadUsesRequestContext(t *testing.T){ctx,cancel:=context.WithCancel(context.Background());cancel();s:=&Service{Repo:&cancelRepo{}};if e:=s.RegisterDownload(ctx,"f");e==nil{t.Fatal("cancellation was ignored")}}

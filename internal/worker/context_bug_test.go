package worker

import (
	"context"
	"testing"
	"time"
)

func TestMergerRunPropagatesCancellation(t *testing.T){ctx,cancel:=context.WithCancel(context.Background());defer cancel();seen:=make(chan bool,1);m:=&Merger{Interval:time.Millisecond,Work:func(c context.Context,_ string)error{seen<-c==ctx;return nil}};go m.Run(ctx);select{case same:=<-seen:if !same{t.Fatal("worker replaced context")};case <-time.After(time.Second):t.Fatal("worker did not run")};cancel()}

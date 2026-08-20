package domain

import "testing"

func TestSortChunksPreservesInput(t *testing.T){in:=[]ChunkRecord{{Index:2},{Index:1}};out:=SortChunks(in);if len(out)!=2||len(in)!=2{t.Fatalf("sort changed slice length: in=%d out=%d",len(in),len(out))}}

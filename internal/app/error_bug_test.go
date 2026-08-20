package app

import (
	"context"
	"errors"
	"testing"
	"chunked-upload-service/internal/store/postgres"
)

func TestEnsurePreservesNotFound(t *testing.T){s:=&Service{Repo:postgres.NewMemoryRepo()};err:=s.Ensure(context.Background(),"missing");if !errors.Is(err,postgres.ErrNotFound){t.Fatalf("errors.Is lost: %v",err)}}

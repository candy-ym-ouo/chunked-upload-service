package config

import "testing"

func TestLoadHasRunnableDefaults(t *testing.T){t.Setenv("UPLOAD_STORAGE_ROOT","");c:=Load();if e:=c.Validate();e!=nil{t.Fatalf("default config invalid: %v",e)}}

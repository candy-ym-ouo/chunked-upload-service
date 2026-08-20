package config

import (
	"chunked-upload-service/internal/domain"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ListenAddr     string
	DatabaseURL    string
	Policy         domain.UploadPolicy
	ReaperInterval time.Duration
	MergerInterval time.Duration
	BearerToken    string
}

func (c Config) Validate() error {
	if c.ListenAddr == "" {
		return fmt.Errorf("listen address required")
	}
	if e := c.Policy.Validate(); e != nil {
		return e
	}
	if c.ReaperInterval <= 0 || c.MergerInterval <= 0 {
		return fmt.Errorf("worker interval must be positive")
	}
	return nil
}
func (c Config) Snapshot() map[string]any {
	return map[string]any{"listen_addr": c.ListenAddr, "database_configured": c.DatabaseURL != "", "storage_root": c.Policy.StorageRoot, "reaper_interval": c.ReaperInterval.String(), "merger_interval": c.MergerInterval.String(), "bearer_enabled": c.BearerToken != ""}
}
func (c Config) Redacted() Config {
	q := c
	if q.BearerToken != "" {
		q.BearerToken = "[redacted]"
	}
	if q.DatabaseURL != "" {
		q.DatabaseURL = "[configured]"
	}
	return q
}
func (c Config) Environment() map[string]string {
	return map[string]string{"UPLOAD_LISTEN_ADDR": c.ListenAddr, "UPLOAD_STORAGE_ROOT": c.Policy.StorageRoot, "UPLOAD_REAPER_INTERVAL": fmt.Sprintf("%d", int(c.ReaperInterval/time.Second)), "UPLOAD_MERGER_INTERVAL": fmt.Sprintf("%d", int(c.MergerInterval/time.Second))}
}
func ParseBool(v string) bool {
	v = strings.ToLower(strings.TrimSpace(v))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}
func ParseSize(v string) int64 {
	v = strings.TrimSpace(strings.ToLower(v))
	mult := int64(1)
	for _, x := range []struct {
		s string
		m int64
	}{{"kb", 1 << 10}, {"mb", 1 << 20}, {"gb", 1 << 30}} {
		if strings.HasSuffix(v, x.s) {
			mult = x.m
			v = strings.TrimSuffix(v, x.s)
			break
		}
	}
	n, _ := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
	return n * mult
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
func duration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if n, e := strconv.Atoi(v); e == nil {
			return time.Duration(n) * time.Second
		}
	}
	return fallback
}
func Load() Config {
	root := env("UPLOAD_STORAGE_ROOT", "./data")
	p := domain.DefaultPolicy(root)
	return Config{ListenAddr: env("UPLOAD_LISTEN_ADDR", ":8080"), DatabaseURL: os.Getenv("DATABASE_URL"), Policy: p, ReaperInterval: duration("UPLOAD_REAPER_INTERVAL", time.Minute), MergerInterval: duration("UPLOAD_MERGER_INTERVAL", 5*time.Second), BearerToken: os.Getenv("UPLOAD_BEARER_TOKEN")}
}

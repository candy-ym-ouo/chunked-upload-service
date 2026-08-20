package http

import (
	"chunked-upload-service/internal/app"
	"chunked-upload-service/internal/domain"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	App         *app.Service
	Log         *slog.Logger
	BearerToken string
}

func New(a *app.Service, l *slog.Logger, tokens ...string) *Server {
	token := ""
	if len(tokens) > 0 {
		token = tokens[0]
	}
	return &Server{App: a, Log: l, BearerToken: token}
}
func (s *Server) Handler() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("/api/v1/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200); w.Write([]byte(`{"status":"ok"}`)) })
	m.HandleFunc("/api/v1/readyz", s.ready)
	m.HandleFunc("/api/v1/metrics", s.metrics)
	m.HandleFunc("/api/v1/uploads", s.uploads)
	m.HandleFunc("/api/v1/uploads/", s.upload)
	m.HandleFunc("/api/v1/files", s.files)
	m.HandleFunc("/api/v1/files/", s.file)
	m.Handle("/", http.FileServer(http.Dir("web")))
	return requestID(s.auth(m))
}
func (s *Server) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.BearerToken != "" && r.URL.Path != "/api/v1/healthz" && r.URL.Path != "/api/v1/readyz" {
			want := "Bearer " + s.BearerToken
			if r.Header.Get("Authorization") != want {
				write(w, 401, map[string]any{"error": map[string]any{"code": "unauthorized", "message": "bearer token required", "request_id": w.Header().Get("X-Request-ID")}})
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
func (s *Server) ready(w http.ResponseWriter, r *http.Request) {
	if e := s.App.Policy.Validate(); e != nil {
		write(w, 503, map[string]string{"status": "not_ready", "error": e.Error()})
		return
	}
	write(w, 200, map[string]string{"status": "ready"})
}
func (s *Server) metrics(w http.ResponseWriter, r *http.Request) {
	c := s.App.Counters
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	fmt.Fprintf(w, "chunked_upload_sessions_created %d\nchunked_upload_chunks_received %d\nchunked_upload_bytes_received %d\nchunked_upload_files_completed %d\nchunked_upload_merge_failures %d\nchunked_upload_downloads %d\n", c.SessionsCreated.Load(), c.ChunksReceived.Load(), c.BytesReceived.Load(), c.FilesCompleted.Load(), c.MergeFailures.Load(), c.Downloads.Load())
}
func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = "req-" + strconv.FormatInt(time.Now().UnixNano(), 10)
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r)
	})
}
func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
func writeError(w http.ResponseWriter, status int, code, message string, details any) {
	write(w, status, map[string]any{"error": map[string]any{"code": code, "message": message, "details": details}})
}
func parseIntQuery(r *http.Request, key string, def int64) (int64, error) {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def, nil
	}
	return strconv.ParseInt(v, 10, 64)
}
func parseLimit(r *http.Request) int {
	v, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if v < 0 {
		return 0
	}
	if v > 1000 {
		return 1000
	}
	return v
}
func acceptsJSON(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept"), "application/json") || r.Header.Get("Accept") == ""
}
func requestMethod(r *http.Request, method string) bool { return strings.EqualFold(r.Method, method) }
func pathParts(v string) []string {
	v = strings.Trim(v, "/")
	if v == "" {
		return nil
	}
	return strings.Split(v, "/")
}
func validIndex(v string) (int, error) {
	i, e := strconv.Atoi(v)
	if e != nil || i < 0 {
		return 0, fmt.Errorf("invalid index")
	}
	return i, nil
}
func headerSize(r *http.Request) (int64, error) {
	if v := r.Header.Get("X-Chunk-Size"); v != "" {
		return strconv.ParseInt(v, 10, 64)
	}
	if r.ContentLength < 0 {
		return 0, fmt.Errorf("content length required")
	}
	return r.ContentLength, nil
}
func setDownloadHeaders(w http.ResponseWriter, f *domain.FileRecord) {
	w.Header().Set("Content-Type", f.MIMEType)
	w.Header().Set("Content-Length", strconv.FormatInt(f.Size, 10))
	w.Header().Set("ETag", fmt.Sprintf(`"%s"`, f.SHA256))
}
func noContent(w http.ResponseWriter) { w.WriteHeader(http.StatusNoContent) }
func methodNotAllowed(w http.ResponseWriter, allowed ...string) {
	w.Header().Set("Allow", strings.Join(allowed, ", "))
	w.WriteHeader(http.StatusMethodNotAllowed)
}
func safeName(v string) string {
	v = strings.TrimSpace(v)
	v = strings.ReplaceAll(v, "/", "_")
	v = strings.ReplaceAll(v, "\\", "_")
	if v == "" {
		return "unnamed"
	}
	return v
}
func contentDisposition(name string) string {
	return fmt.Sprintf(`attachment; filename="%s"`, safeName(name))
}
func (s *Server) uploads(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.NotFound(w, r)
		return
	}
	var q struct {
		UploadKey        string `json:"upload_key"`
		FileName         string `json:"file_name"`
		MIMEType         string `json:"mime_type"`
		ExpectedChecksum string `json:"expected_checksum"`
		TotalSize        int64  `json:"total_size"`
		ChunkSize        int64  `json:"chunk_size"`
	}
	if r.Body == nil || json.NewDecoder(io.LimitReader(r.Body, 64<<10)).Decode(&q) != nil {
		write(w, 400, map[string]string{"error": "invalid json"})
		return
	}
	u, exists, e := s.App.Create(r.Context(), q.UploadKey, q.FileName, q.MIMEType, q.TotalSize, q.ChunkSize, q.ExpectedChecksum)
	if e != nil {
		write(w, 422, map[string]string{"error": e.Error()})
		return
	}
	if exists {
		write(w, 200, u)
	} else {
		write(w, 201, u)
	}
}
func (s *Server) upload(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		http.NotFound(w, r)
		return
	}
	id := parts[3]
	if len(parts) >= 6 && parts[4] == "chunks" {
		i, e := validIndex(parts[5])
		if e != nil {
			writeError(w, http.StatusUnprocessableEntity, "invalid_chunk_index", e.Error(), nil)
			return
		}
		if r.Method != "PUT" {
			http.NotFound(w, r)
			return
		}
		size := r.ContentLength
		if h := r.Header.Get("X-Chunk-Size"); h != "" {
			size, _ = strconv.ParseInt(h, 10, 64)
		}
		if size <= 0 {
			write(w, 422, map[string]string{"error": "missing content length"})
			return
		}
		if r.Header.Get("X-Chunk-SHA256") != "" { /* application recomputes and persists the authoritative digest */
		}
		c, e := s.App.PutChunk(r.Context(), id, i, io.LimitReader(r.Body, size+1), size)
		if e != nil {
			write(w, 422, map[string]string{"error": e.Error()})
			return
		}
		write(w, 200, c)
		return
	}
	if parts[4] == "complete" && r.Method == "POST" {
		f, e := s.App.Complete(r.Context(), id)
		if e != nil {
			code := 422
			if strings.Contains(e.Error(), "checksum mismatch") || strings.Contains(e.Error(), "merge in progress") {
				code = http.StatusConflict
			}
			write(w, code, map[string]string{"error": e.Error()})
			return
		}
		write(w, 200, f)
		return
	}
	if r.Method == "GET" {
		u, c, e := s.App.Session(r.Context(), id)
		if e != nil {
			write(w, 404, map[string]string{"error": "not found"})
			return
		}
		write(w, 200, map[string]any{"session": u, "resume": app.BuildResume(u, c)})
		return
	}
	if r.Method == "DELETE" {
		e := s.App.Abort(r.Context(), id)
		if e != nil {
			write(w, 404, map[string]string{"error": e.Error()})
			return
		}
		w.WriteHeader(204)
		return
	}
	http.NotFound(w, r)
}
func (s *Server) files(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		q := r.URL.Query()
		limit, _ := strconv.Atoi(q.Get("limit"))
		if limit < 0 {
			limit = 0
		}
		min, _ := strconv.ParseInt(q.Get("min_size"), 10, 64)
		max, _ := strconv.ParseInt(q.Get("max_size"), 10, 64)
		var before time.Time
		if v := q.Get("before"); v != "" {
			before, _ = time.Parse(time.RFC3339, v)
		}
		f, e := s.App.QueryFiles(r.Context(), app.FileQuery{NamePrefix: q.Get("name"), MIME: q.Get("mime"), MinSize: min, MaxSize: max, Before: before, Limit: limit})
		if e != nil {
			write(w, 500, map[string]string{"error": e.Error()})
			return
		}
		write(w, 200, f)
		return
	}
	http.NotFound(w, r)
}
func (s *Server) file(w http.ResponseWriter, r *http.Request) {
	id := path.Base(r.URL.Path)
	content := strings.HasSuffix(r.URL.Path, "/content")
	if content {
		id = path.Base(path.Dir(r.URL.Path))
	}
	f, e := s.App.File(r.Context(), id)
	if e != nil {
		write(w, 404, map[string]string{"error": "not found"})
		return
	}
	if content {
		rs, e := s.App.Storage.OpenFile(r.Context(), f.StoredPath)
		if e != nil {
			write(w, 404, map[string]string{"error": "content missing"})
			return
		}
		defer rs.Close()
		w.Header().Set("ETag", fmt.Sprintf(`"%s"`, f.SHA256))
		if strings.Trim(r.Header.Get("If-None-Match"), `"`) == f.SHA256 {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		_ = s.App.RegisterDownload(r.Context(), f.ID)
		s.App.Counters.Downloads.Add(1)
		http.ServeContent(w, r, f.Name, f.UploadedAt, rs)
		return
	}
	if r.Method == "DELETE" {
		if e = s.App.DeleteFile(r.Context(), id); e != nil {
			write(w, 500, map[string]string{"error": e.Error()})
			return
		}
		w.WriteHeader(204)
		return
	}
	write(w, 200, f)
}

package domain

import "time"

type FileRecord struct {
	ID            string     `json:"id"`
	SessionID     string     `json:"session_id"`
	Name          string     `json:"name"`
	Size          int64      `json:"size"`
	SHA256        string     `json:"sha256"`
	ChunkCount    int        `json:"chunk_count"`
	MIMEType      string     `json:"mime_type"`
	Status        string     `json:"status"`
	DownloadCount int64      `json:"download_count"`
	UploadedAt    time.Time  `json:"uploaded_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty"`
	StoredPath    string     `json:"-"`
}

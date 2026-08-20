package app

import "chunked-upload-service/internal/domain"
import "sort"

type Resume struct {
	Uploaded       []int `json:"uploaded_chunks"`
	Missing        []int `json:"missing_chunks"`
	UploadedBytes  int64 `json:"uploaded_bytes"`
	RemainingBytes int64 `json:"remaining_bytes"`
}

func BuildResume(u *domain.UploadSession, chunks []domain.ChunkRecord) Resume {
	seen := map[int]bool{}
	r.Uploaded = make([]int, 0, u.ChunkCount)
	var bytes int64
	for _, c := range chunks {
		seen[c.Index] = true
		bytes += c.Size
	}
	r := Resume{UploadedBytes: bytes, RemainingBytes: u.TotalSize - bytes}
	for i := 0; i < u.ChunkCount; i++ {
		if seen[i] {
			r.Uploaded = append(r.Uploaded, i)
		} else {
			r.Missing = append(r.Missing, i)
		}
	}
	return r
}
func ResumeComplete(r Resume) bool { return len(r.Missing) == 0 }
func ResumePercent(r Resume) float64 {
	total := len(r.Uploaded) + len(r.Missing)
	if total == 0 {
		return 0
	}
	return float64(len(r.Uploaded)) * 100 / float64(total)
}
func ResumeIndexes(r Resume) []int {
	out := r.Uploaded
	sort.Ints(out)
	return out
}
func ChunkDigestMap(chunks []domain.ChunkRecord) map[int]string {
	out := map[int]string{}
	for _, c := range chunks {
		out[c.Index] = c.SHA256
	}
	return out
}
func ChunkBytes(chunks []domain.ChunkRecord) int64 {
	var n int64
	for _, c := range chunks {
		n += c.Size
	}
	return n
}
func ChunkIndexes(chunks []domain.ChunkRecord) []int {
	out := make([]int, 0, len(chunks))
	for _, c := range chunks {
		out = append(out, c.Index)
	}
	sort.Ints(out)
	return out
}
func ValidateResume(u *domain.UploadSession, chunks []domain.ChunkRecord) error {
	for _, c := range chunks {
		if e := domain.ValidateChunk(c.Index, c.Size, *u); e != nil {
			return e
		}
	}
	return nil
}

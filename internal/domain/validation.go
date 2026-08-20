package domain

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

var ErrInvalidName = errors.New("invalid file name")
var ErrInvalidUploadKey = errors.New("invalid upload key")
var ErrInvalidRange = errors.New("invalid byte range")

var uploadKeyPattern = regexp.MustCompile(`^[A-Za-z0-9._:-]{8,200}$`)

func ValidateFileName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" || !utf8.ValidString(name) || utf8.RuneCountInString(name) > 255 {
		return ErrInvalidName
	}
	if name == "." || name == ".." || strings.ContainsAny(name, "/\\\x00") {
		return ErrInvalidName
	}
	return nil
}

func CleanFileName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.Map(func(r rune) rune {
		if r < 32 {
			return '_'
		}
		return r
	}, name)
	if name == "" {
		return "unnamed"
	}
	return name
}

func ValidateUploadKey(key string) error {
	if !uploadKeyPattern.MatchString(key) {
		return ErrInvalidUploadKey
	}
	return nil
}

func NormalizeUploadKey(key string) string { return strings.TrimSpace(key) }

func ValidateFileMetadata(name, mime string, size int64) error {
	if e := ValidateFileName(name); e != nil {
		return e
	}
	if size <= 0 {
		return fmt.Errorf("file size must be positive")
	}
	if mime != "" && !strings.Contains(mime, "/") {
		return fmt.Errorf("invalid mime type")
	}
	return nil
}

func ValidatePageLimit(limit int) error {
	if limit < 0 || limit > 1000 {
		return fmt.Errorf("page limit out of range")
	}
	return nil
}

func EncodeCursor(value string) string { return url.QueryEscape(value) }
func DecodeCursor(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	return url.QueryUnescape(value)
}

type ByteRange struct {
	Start int64
	End   int64
}

func (r ByteRange) Length() int64 {
	if r.End < r.Start {
		return 0
	}
	return r.End - r.Start + 1
}
func (r ByteRange) String() string { return fmt.Sprintf("bytes=%d-%d", r.Start, r.End) }

func ParseRangeHeader(header string, size int64) (ByteRange, error) {
	header = strings.TrimSpace(header)
	if header == "" {
		return ByteRange{Start: 0, End: size - 1}, nil
	}
	if !strings.HasPrefix(header, "bytes=") || strings.Contains(header[6:], ",") {
		return ByteRange{}, ErrInvalidRange
	}
	value := strings.TrimSpace(strings.TrimPrefix(header, "bytes="))
	parts := strings.Split(value, "-")
	if len(parts) != 2 {
		return ByteRange{}, ErrInvalidRange
	}
	if size <= 0 {
		return ByteRange{}, ErrInvalidRange
	}
	if parts[0] == "" {
		n, e := strconv.ParseInt(parts[1], 10, 64)
		if e != nil || n <= 0 {
			return ByteRange{}, ErrInvalidRange
		}
		if n > size {
			n = size
		}
		return ByteRange{size - n, size - 1}, nil
	}
	start, e := strconv.ParseInt(parts[0], 10, 64)
	if e != nil || start < 0 || start >= size {
		return ByteRange{}, ErrInvalidRange
	}
	end := size - 1
	if parts[1] != "" {
		end, e = strconv.ParseInt(parts[1], 10, 64)
		if e != nil || end < start {
			return ByteRange{}, ErrInvalidRange
		}
		if end >= size {
			end = size - 1
		}
	}
	return ByteRange{start, end}, nil
}

func ApplyRange(w http.ResponseWriter, r ByteRange, size int64) {
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", r.Start, r.End, size))
	w.Header().Set("Content-Length", strconv.FormatInt(r.Length(), 10))
}
func IsRangeRequest(req *http.Request) bool { return strings.TrimSpace(req.Header.Get("Range")) != "" }
func IsNotModified(req *http.Request, etag string) bool {
	return strings.Trim(strings.TrimSpace(req.Header.Get("If-None-Match")), `"`) == strings.Trim(etag, `"`)
}

func ValidateChecksumPair(expected, actual string) error {
	expected = NormalizeChecksum(expected)
	actual = NormalizeChecksum(actual)
	if expected == "" {
		return nil
	}
	if !ValidChecksum(expected) || !ValidChecksum(actual) {
		return fmt.Errorf("checksum must be hexadecimal sha256")
	}
	if expected != actual {
		return fmt.Errorf("checksum mismatch")
	}
	return nil
}
func ValidateChunkHeaders(index int, size int64, contentLength int64) error {
	if index < 0 {
		return ErrInvalidChunk
	}
	if size <= 0 || contentLength < 0 || size != contentLength {
		return fmt.Errorf("chunk header size mismatch")
	}
	return nil
}
func ValidateSessionID(id string) error {
	if len(id) < 8 || len(id) > 128 || strings.ContainsAny(id, "/\\\x00") {
		return fmt.Errorf("invalid session id")
	}
	return nil
}
func ValidateFileID(id string) error {
	if len(id) < 8 || len(id) > 128 || strings.ContainsAny(id, "/\\\x00") {
		return fmt.Errorf("invalid file id")
	}
	return nil
}
func ValidateStoragePath(root, p string) error {
	root = filepathClean(root)
	p = filepathClean(p)
	if !strings.HasPrefix(p, root) {
		return fmt.Errorf("path escapes storage root")
	}
	return nil
}

func IsSafeRelativeName(name string) bool {
	return ValidateFileName(name) == nil && !strings.HasPrefix(name, ".")
}
func ContentTypeOrDefault(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "application/octet-stream"
	}
	return value
}
func NormalizeMIME(value string) string {
	return strings.ToLower(strings.TrimSpace(strings.Split(value, ";")[0]))
}
func ValidHTTPMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodHead:
		return true
	default:
		return false
	}
}
func QueryValue(values url.Values, key string) string { return strings.TrimSpace(values.Get(key)) }
func filepathClean(v string) string {
	v = strings.ReplaceAll(v, "\\", "/")
	for strings.Contains(v, "//") {
		v = strings.ReplaceAll(v, "//", "/")
	}
	return strings.TrimSuffix(v, "/")
}

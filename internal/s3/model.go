package s3

import (
	"io"
	"time"
)

type ObjectMetadataResponse struct {
	Key      string   `json:"key"`
	Metadata Metadata `json:"metadata"`
}

type Metadata struct {
	Size         int64     `json:"size"`
	ContentType  string    `json:"content_type"`
	LastModified time.Time `json:"last_modified"`
}

type ObjectListResponse struct {
	Prefix      string  `json:"prefix"`
	ObjectCount int     `json:"object_count"`
	TotalSize   int64   `json:"total_size"`
	Objects     Objects `json:"objects"`
}

type Object struct {
	Key          string    `json:"key"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"last_modified"`
}

type Objects []Object

func (o Objects) Keys() []string {
	keys := make([]string, 0)
	for _, obj := range o {
		keys = append(keys, obj.Key)
	}
	return keys
}

type writerWrapper struct {
	w io.Writer
}

func NewWriterWrapper(w io.Writer) (io.WriterAt, error) {
	if w == nil {
		return nil, ErrNilWriter
	}
	return &writerWrapper{w: w}, nil
}

func (w *writerWrapper) WriteAt(p []byte, off int64) (n int, err error) {
	return w.w.Write(p)
}

type Tag struct {
	Key   string
	Value string
}

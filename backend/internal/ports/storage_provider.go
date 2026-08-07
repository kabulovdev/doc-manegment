package ports

import (
	"context"
	"io"
)

type ObjectInfo struct {
	Key          string
	Size         int64
	ETag         string
	ContentType  string
	LastModified int64
}

type Range struct {
	Start int64
	End   int64 // inclusive; -1 means through the end
}

type ListPage struct {
	Objects     []ObjectInfo
	TotalBytes  int64
	TotalCount  int64
	NextCursor  string
	IsTruncated bool
}

type Part struct {
	PartNumber int
	ETag       string
}

type StorageProvider interface {
	Stat(ctx context.Context, key string) (ObjectInfo, error)
	GetStream(ctx context.Context, key string, rng *Range) (io.ReadCloser, ObjectInfo, error)
	PutStream(ctx context.Context, key string, body io.Reader, size int64, mime string) error
	Delete(ctx context.Context, key string) error
	List(ctx context.Context, prefix, cursor string) (ListPage, error)
	InitiateMultipart(ctx context.Context, key, mime string) (uploadID string, err error)
	UploadPart(ctx context.Context, key, uploadID string, partNum int, body io.Reader, size int64) (etag string, err error)
	CompleteMultipart(ctx context.Context, key, uploadID string, parts []Part) error
	AbortMultipart(ctx context.Context, key, uploadID string) error
	HeadBucket(ctx context.Context) error
}

package gcs

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

type RWHandle struct {
	Bucket *storage.BucketHandle
 	Reader *storage.Reader
	Writer *storage.Writer
	Client *storage.Client
}

// IMPORTANT - if you don't call this (and see it return nil), your data is likely lost
func (h *RWHandle)Close() error {
	if h.Reader != nil {
		h.Reader.Close()
	}

	if h.Writer != nil {
		if err := h.Writer.Close(); err != nil {
			return err
		}
	}

	if err := h.Client.Close(); err != nil {
		return err
	}

	return nil
}

func (h *RWHandle)IOReader() io.Reader {
	return io.Reader(h.Reader)
}
func (h *RWHandle)IOWriter() io.Writer {
	return io.Writer(h.Writer)
}

func Exists(ctx context.Context, bucketname string, filename string) (bool,error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		//log.Errorf(ctx, "failed to get a client: %v", err)
		return false, err
	}
	
	bucket := client.Bucket(bucketname)
	if bucket == nil {
		return false, fmt.Errorf("GCS client.Bucket() was nil")
	}

	if _,err := bucket.Object(filename).NewReader(ctx); err == nil {
		return true,nil
	} else if err == storage.ErrObjectNotExist {
		return false,nil
	} else {
		return false,err
	}
}

// OpenBucket returns a handle that can read a bucket, but the bucket must already exist.
func ListBucket(ctx context.Context, bucketname string) ([]string, error) {
	contents := []string{}

	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	bucket := client.Bucket(bucketname)
	it := bucket.Objects(ctx, &storage.Query{Prefix: ""})
	for {
    attrs, err := it.Next()
    if err == iterator.Done {
			break
    }
    if err != nil {
			return []string{}, fmt.Errorf("bucket list iterator(%s): %v\n", bucketname, err)
    }
    contents = append(contents, attrs.Name)
	}

	return contents, nil
}

// This appears to truncate contents :/
func OpenRW(ctx context.Context, bucketname string, filename string, contentType string) (*RWHandle, error) {
	handle := RWHandle{}
	if c, err := storage.NewClient(ctx); err != nil {
		//log.Errorf(ctx, "failed to get a client: %v", err)
		return nil, err
	} else {
		handle.Client = c
	}
	
	bucket := handle.Client.Bucket(bucketname)
	if bucket == nil {
		return nil, fmt.Errorf("GCS client.Bucket() was nil")
	}

	// NOTE - this may be nil, if file does not yet exist
	rdr, err := bucket.Object(filename).NewReader(ctx)
	handle.Reader = rdr
	_=err
	//if err != nil {
	//	return nil, fmt.Errorf("bucket=%s,file=%s./NewReader: %v\n", bucketname, filename, err)
	//}

	handle.Writer = bucket.Object(filename).NewWriter(ctx)
	handle.Writer.ContentType = contentType //"text/plain" // CSV?
	
	return &handle, nil
}


func OpenR(ctx context.Context, bucketname string, filename string) (*RWHandle, error) {
	handle := RWHandle{}
	if c, err := storage.NewClient(ctx); err != nil {
		return nil, err
	} else {
		handle.Client = c
	}

	bucket := handle.Client.Bucket(bucketname)
	if bucket == nil {
		return nil, fmt.Errorf("GCS client.Bucket() was nil")
	}

	// NOTE - this may be nil, if file does not yet exist
	rdr, err := bucket.Object(filename).NewReader(ctx)
	handle.Reader = rdr
	if err != nil {
		return nil, fmt.Errorf("gcs.OpenR: bucket=%s,file=%s./NewReader: %v\n", bucketname, filename, err)
	}

	return &handle, nil
}


func (h *RWHandle)ToReader(ctx context.Context, bucketname, filename string) (io.Reader, error) {
	bucket := h.Client.Bucket(bucketname)
	if bucket == nil {
		return nil, fmt.Errorf("GCS client.Bucket() was nil")
	}

	r,err := bucket.Object(filename).NewReader(ctx)
	if err != nil {
		return nil, err
	}

	return io.Reader(r), nil
}

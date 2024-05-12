// Package uploadgcs is an upload file store for Google Cloud Storage
package uploadgcs

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
	"go.uber.org/zap"
)

type Store struct {
	log    *zap.SugaredLogger
	client *storage.Client
	bucket string
}

func NewStore(log *zap.SugaredLogger, bucket string) (*Store, error) {
	client, err := storage.NewClient(context.Background())
	if err != nil {
		return nil, err
	}

	s := &Store{
		log:    log,
		client: client,
		bucket: bucket,
	}
	return s, nil
}

// Save copies the source reader contents to a new file in the bucket defined
// within the Store and the name provided in the function as the full path
func (s *Store) Save(name string, src io.ReadCloser) error {
	ctx := context.Background()
	bkt := s.client.Bucket(s.bucket)
	obj := bkt.Object(name)
	dst := obj.NewWriter(ctx)
	defer dst.Close()
	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}
	return nil
}

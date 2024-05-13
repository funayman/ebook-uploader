// Package uploads3 is an upload file store for AWS S3
package uploads3

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
)

type Store struct {
	log    *zap.SugaredLogger
	client *s3.Client
	bucket string
}

func NewStore(log *zap.SugaredLogger, bucket string) (*Store, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	config, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("config.LoadDefault: %w", err)
	}
	return NewStoreFromConfig(log, bucket, config)
}

func NewStoreFromConfig(log *zap.SugaredLogger, bucket string, config aws.Config) (*Store, error) {
	client := s3.NewFromConfig(config)
	return &Store{log: log, client: client, bucket: bucket}, nil
}

func (s Store) Save(name string, src io.ReadCloser) error {
	ctx := context.Background()

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &s.bucket,
		Key:    &name,
		Body:   src,
	})

	if err != nil {
		return err
	}

	return nil
}

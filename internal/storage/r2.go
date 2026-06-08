package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type R2Store struct {
	client    *s3.Client
	bucket    string
	publicURL string // base URL for public access, e.g. https://pub-xxx.r2.dev
}

func NewR2Store(accountID, accessKey, secretKey, bucket, publicURL string) *R2Store {
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)
	client := s3.New(s3.Options{
		BaseEndpoint: aws.String(endpoint),
		Region:       "auto",
		Credentials:  credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
	})
	return &R2Store{client: client, bucket: bucket, publicURL: publicURL}
}

func (s *R2Store) Upload(ctx context.Context, key string, r io.Reader, contentType string) (string, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        r,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", s.publicURL, key), nil
}

func (s *R2Store) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

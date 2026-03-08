package storages

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type s3Storage struct {
	cfg     Config
	client  *s3.Client
	presign *s3.PresignClient
}

func newS3Storage(cfg Config) (Storage, error) {
	resolver := s3.EndpointResolverFromURL(cfg.Endpoint)

	loadOptions := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.Region),
	}

	if strings.TrimSpace(cfg.AccessKeyID) != "" && strings.TrimSpace(cfg.SecretAccessKey) != "" {
		loadOptions = append(loadOptions, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, cfg.SessionToken),
		))
	}

	if strings.TrimSpace(cfg.Endpoint) != "" {
		loadOptions = append(loadOptions, awsconfig.WithBaseEndpoint(cfg.Endpoint))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(), loadOptions...)
	if err != nil {
		return nil, err
	}

	clientOptions := func(o *s3.Options) {
		o.UsePathStyle = cfg.UsePathStyle
		if strings.TrimSpace(cfg.Endpoint) != "" {
			o.EndpointResolver = resolver
		}
	}

	client := s3.NewFromConfig(awsCfg, clientOptions)
	return &s3Storage{
		cfg:     cfg,
		client:  client,
		presign: s3.NewPresignClient(client),
	}, nil
}

func (s *s3Storage) PutObject(ctx context.Context, key string, body io.Reader, contentType string, metadata map[string]string) error {
	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" {
		return ErrObjectKeyRequired
	}

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.cfg.Bucket),
		Key:         aws.String(trimmedKey),
		Body:        body,
		ContentType: aws.String(strings.TrimSpace(contentType)),
		Metadata:    metadata,
	})
	return err
}

func (s *s3Storage) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" {
		return nil, ErrObjectKeyRequired
	}

	res, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.cfg.Bucket),
		Key:    aws.String(trimmedKey),
	})
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}

func (s *s3Storage) HeadObject(ctx context.Context, key string) (ObjectInfo, error) {
	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" {
		return ObjectInfo{}, ErrObjectKeyRequired
	}

	res, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.cfg.Bucket),
		Key:    aws.String(trimmedKey),
	})
	if err != nil {
		return ObjectInfo{}, err
	}

	info := ObjectInfo{
		Key:         trimmedKey,
		Size:        aws.ToInt64(res.ContentLength),
		ContentType: aws.ToString(res.ContentType),
		ETag:        strings.Trim(aws.ToString(res.ETag), "\""),
		Metadata:    res.Metadata,
	}
	if res.LastModified != nil {
		info.LastModified = *res.LastModified
	}

	return info, nil
}

func (s *s3Storage) DeleteObject(ctx context.Context, key string) error {
	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" {
		return ErrObjectKeyRequired
	}

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.cfg.Bucket),
		Key:    aws.String(trimmedKey),
	})
	return err
}

func (s *s3Storage) PresignGetURL(ctx context.Context, key string, expiresIn time.Duration) (string, error) {
	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" {
		return "", ErrObjectKeyRequired
	}
	if expiresIn <= 0 {
		expiresIn = 15 * time.Minute
	}

	res, err := s.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.cfg.Bucket),
		Key:    aws.String(trimmedKey),
	}, s3.WithPresignExpires(expiresIn))
	if err != nil {
		return "", err
	}

	return res.URL, nil
}

func (s *s3Storage) PresignPutURL(ctx context.Context, key string, expiresIn time.Duration, contentType string) (string, error) {
	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" {
		return "", ErrObjectKeyRequired
	}
	if expiresIn <= 0 {
		expiresIn = 15 * time.Minute
	}

	res, err := s.presign.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.cfg.Bucket),
		Key:         aws.String(trimmedKey),
		ContentType: aws.String(strings.TrimSpace(contentType)),
	}, s3.WithPresignExpires(expiresIn))
	if err != nil {
		return "", err
	}

	return res.URL, nil
}

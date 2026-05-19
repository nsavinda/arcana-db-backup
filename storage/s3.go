package storage

import (
	"context"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Config struct {
	Bucket    string
	Region    string
	AccessKey string
	SecretKey string
	Endpoint  string
}

func newClient(ctx context.Context, cfg S3Config) (*s3.Client, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKey, cfg.SecretKey, "")),
	)
	if err != nil {
		return nil, err
	}

	return s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
		o.UsePathStyle = true
	}), nil
}

// Upload uploads a file to S3/DigitalOcean Spaces
func Upload(cfg S3Config, filename string) error {
	ctx := context.Background()

	client, err := newClient(ctx, cfg)
	if err != nil {
		return err
	}

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(cfg.Bucket),
		Key:           aws.String(filename),
		Body:          file,
		ContentLength: aws.Int64(fileInfo.Size()),
		ContentType:   aws.String("application/octet-stream"),
	})
	return err
}

// Download fetches an object by key and writes it to destPath.
func Download(cfg S3Config, key, destPath string) error {
	ctx := context.Background()

	client, err := newClient(ctx, cfg)
	if err != nil {
		return err
	}

	out, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(cfg.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}
	defer out.Body.Close()

	f, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, out.Body)
	return err
}

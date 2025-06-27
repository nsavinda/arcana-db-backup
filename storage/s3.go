package storage

import (
	"bytes"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3Config struct {
	Bucket    string
	Region    string
	AccessKey string
	SecretKey string
	Endpoint  string
}

// Upload uploads a file to S3/DigitalOcean Spaces
func Upload(cfg S3Config, filename string) error {
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(cfg.Region),
		Endpoint:         aws.String(cfg.Endpoint),
		S3ForcePathStyle: aws.Bool(true),
		Credentials: credentials.NewStaticCredentials(
			cfg.AccessKey, cfg.SecretKey, ""),
	})
	if err != nil {
		return err
	}
	svc := s3.New(sess)

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	size := fileInfo.Size()
	buffer := make([]byte, size)
	file.Read(buffer)

	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(cfg.Bucket),
		Key:           aws.String(filename),
		Body:          bytes.NewReader(buffer),
		ContentLength: aws.Int64(size),
		ContentType:   aws.String("application/octet-stream"),
	})
	return err
}

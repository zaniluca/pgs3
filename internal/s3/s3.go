package s3

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3Client struct {
	*s3.S3
}

func NewClient(accessKeyId, secretAccessKey, region, endpoint string) (*S3Client, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(region),
		Endpoint:         aws.String(endpoint),
		S3ForcePathStyle: aws.Bool(true),
		Credentials: credentials.NewStaticCredentials(
			accessKeyId,
			secretAccessKey,
			"", // Token can be empty for this use case
		),
	})

	return &S3Client{s3.New(sess)}, err
}

func (c S3Client) UploadFile(bucket, file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = c.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filepath.Base(file)),
		Body:   f,
	})

	return err
}

func (c S3Client) RemoveOldBackups(bucket string, before time.Time) error {
	fmt.Printf("Removing backups older than %s\n", before)
	resp, err := c.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(bucket)})
	if err != nil {
		return err
	}

	for _, item := range resp.Contents {
		if item.LastModified.Before(before) {
			_, err := c.DeleteObject(&s3.DeleteObjectInput{
				Bucket: aws.String(bucket),
				Key:    item.Key,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Deleted: %s\n", *item.Key)
		}
	}

	return nil
}

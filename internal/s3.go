package pgs3

import (
	"fmt"
	"io"
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

func NewS3Client(accessKeyId, secretAccessKey, region, endpoint string) (*S3Client, error) {
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

func (c S3Client) DownloadFile(bucket, key, file string) (*os.File, error) {
	localFile, err := os.Create(file)
	if err != nil {
		return nil, err
	}
	defer localFile.Close()

	resp, err := c.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	_, err = io.Copy(localFile, resp.Body)
	return localFile, err
}

func (c S3Client) LatestBackup(bucket string) (string, error) {
	resp, err := c.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(bucket)})
	if err != nil {
		return "", err
	}

	if len(resp.Contents) == 0 {
		return "", fmt.Errorf("no backups found in bucket %s", bucket)
	}

	latest := resp.Contents[0]
	for _, item := range resp.Contents {
		if item.LastModified.After(*latest.LastModified) {
			latest = item
		}
	}

	return *latest.Key, nil
}

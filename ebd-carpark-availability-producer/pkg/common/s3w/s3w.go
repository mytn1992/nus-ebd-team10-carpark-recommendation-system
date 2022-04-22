package s3w

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type S3 interface {
	PutObject(bucket string, objectkey string, filepath string) error
	ListObject(bucket string, objectkey string, filename string) error
	CopyObject(source string, destination string, objectkey string) error
	DeleteObject(bucket string, objectkey string) error
}

type Wrapper struct {
	s3svc           *s3.S3
	s3Uploadersvc   *s3manager.Uploader
	s3Downloadersvc *s3manager.Downloader
}

func (w *Wrapper) PutObject(bucket string, objectkey string, filepath string) error {
	filename := filepath
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file %q, %v", filename, err)
	}

	_, err = w.s3Uploadersvc.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectkey),
		Body:   f,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file, %v", err)
	}
	return nil
}

func (w *Wrapper) ListObject(bucket string, objectkey string, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %q, %v", filename, err)
	}

	_, err = w.s3Downloadersvc.Download(f, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectkey),
	})
	if err != nil {
		return fmt.Errorf("failed to download file, %v", err)
	}
	return nil
}

func (w *Wrapper) CopyObject(source string, destination string, objectkey string) error {
	input := &s3.CopyObjectInput{
		CopySource: aws.String(source),
		Bucket:     aws.String(destination),
		Key:        aws.String(objectkey),
	}

	_, err := w.s3svc.CopyObject(input)

	if err != nil {
		return err
	}
	return nil
}

func (w *Wrapper) DeleteObject(bucket string, objectkey string) error {
	_, err := w.s3svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectkey),
	})
	if err != nil {
		return err
	}
	return nil
}

func NewWrapper(sess *session.Session) S3 {
	return &Wrapper{
		s3svc:           s3.New(sess),
		s3Uploadersvc:   s3manager.NewUploader(sess),
		s3Downloadersvc: s3manager.NewDownloader(sess),
	}
}

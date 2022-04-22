package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	push_processed_s3 "github.com/mytn1992/ebd-carpark-availability-producer/pkg/cmd/push-processed-s3"
	"github.com/mytn1992/ebd-carpark-availability-producer/pkg/common/s3w"
	log "github.com/sirupsen/logrus"
)

func Handler(ctx context.Context, S3Event events.S3Event) {
	bucketName := S3Event.Records[0].S3.Bucket.Name
	objectKey := S3Event.Records[0].S3.Object.Key
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("ap-southeast-1"),
	})
	if err != nil {
		log.Errorf("can't init session %v", err)
	}
	s3wrapper := s3w.NewWrapper(sess)

	//download file from s3
	err = s3wrapper.ListObject(bucketName, objectKey, fmt.Sprintf("/tmp/%v", strings.Split(objectKey, "/")[1]))
	if err != nil {
		log.Fatal(err)
	}
	//process record
	_, _, err = push_processed_s3.Run(fmt.Sprintf("/tmp/%v", strings.Split(objectKey, "/")[1]))
	if err != nil {
		log.Fatal(err)
	}
	log.Info("Successfully pushed data to es")

	//copy file to processed folder
	err = s3wrapper.CopyObject("/"+bucketName+"/"+objectKey, bucketName, "processed/"+strings.Split(objectKey, "/")[1])
	if err != nil {
		log.Fatal(err)
	}
	//delete file in input folder
	err = s3wrapper.DeleteObject(bucketName, objectKey)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	lambda.Start(Handler)
}

func GenFileName() string {
	return fmt.Sprintf("carpark-all-location-%v.csv", time.Now().Add(8*time.Hour).Format("20060102T15"))
}

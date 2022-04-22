package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/mytn1992/ebd-carpark-availability-producer/pkg/cmd/extract"
	"github.com/mytn1992/ebd-carpark-availability-producer/pkg/common/lambdaUtil"
)

type Lambda struct {
}

func (l Lambda) HandleRequest(ctx context.Context, req map[string]interface{}) (interface{}, error) {
	res, statusCode, err := extract.Run()
	return lambdaUtil.BuildLambdaResponse(res, statusCode, err)
}

func main() {
	l := Lambda{}
	if os.Getenv("AWS_EXECUTION_ENV") != "" {
		lambda.Start(l.HandleRequest)
	} else {
		l.HandleRequest(context.Background(), map[string]interface{}{})
	}
}

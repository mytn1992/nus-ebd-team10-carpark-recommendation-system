package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/mytn1992/ebd-carpark-streaming-api/pkg/cmd/random_elastic"
	"github.com/mytn1992/ebd-carpark-streaming-api/pkg/cmd/streamingapi"
	"github.com/mytn1992/ebd-carpark-streaming-api/pkg/common/lambdaUtil"
)

// Handler function Using AWS Lambda Proxy Request
func Handler(request events.APIGatewayProxyRequest) (interface{}, error) {

	// BodyRequest will be used to take the json response from client and build it
	bodyRequest := random_elastic.Buildings{}

	// Unmarshal the json, return 404 if error
	err := json.Unmarshal([]byte(request.Body), &bodyRequest)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 404}, nil
	}

	result, statusCode, err := streamingapi.Run(bodyRequest)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: statusCode}, nil
	}
	fmt.Println(len(result))
	//Returning response with AWS Lambda Proxy Response
	return lambdaUtil.BuildApigatewayJSONResponse(result, statusCode, err, lambdaUtil.BuildHeader("*"))
}

func main() {
	if os.Getenv("AWS_EXECUTION_ENV") != "" {
		lambda.Start(Handler)
	} else {
		input := `{
			"input": "changi",
			"distance": 1000
		}`
		body := random_elastic.Buildings{}
		err := json.Unmarshal([]byte(input), &body)
		if err != nil {
			log.Fatal(err)
		}
		result, _, err := streamingapi.Run(body)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(len(result))
	}
}

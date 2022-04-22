package main

import (
	"encoding/json"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/mytn1992/ebd-carpark-streaming-api/pkg/cmd/random_elastic"
)

// BodyRequest is our self-made struct to process JSON request from Client
type BodyRequest struct {
	Input string `json:"input"`
}

// BodyResponse is our self-made struct to build response for Client
type BodyResponse struct {
	ResponseName string `json:"name"`
}

// Handler function Using AWS Lambda Proxy Request
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	// BodyRequest will be used to take the json response from client and build it
	bodyRequest := BodyRequest{}

	// Unmarshal the json, return 404 if error
	err := json.Unmarshal([]byte(request.Body), &bodyRequest)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 404}, nil
	}

	result, statusCode, err := random_elastic.Run()
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: statusCode}, nil
	}

	//Returning response with AWS Lambda Proxy Response
	return events.APIGatewayProxyResponse{Body: result, StatusCode: statusCode}, nil
}

func main() {
	if os.Getenv("AWS_EXECUTION_ENV") != "" {
		lambda.Start(Handler)
	} else {
		random_elastic.Run()
	}
}

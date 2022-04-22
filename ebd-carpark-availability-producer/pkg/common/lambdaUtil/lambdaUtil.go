package lambdaUtil

import (
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

const (
	X_Slack_No_Retry  = "X-Slack-No-Retry"
	X_Slack_Retry_Num = "X-Slack-Retry-Num"
)

type StandardResponse struct {
	StatusCode int         `json:"statusCode"`
	Message    string      `json:"message"`
	Body       interface{} `json:"body"`
}

func BuildLambdaResponse(body interface{}, statusCode int, err error) (StandardResponse, error) {
	return StandardResponse{
		StatusCode: statusCode,
		Body:       body,
	}, err
}

func BuildApigatewayResponse(body interface{}, statusCode int, err error) (events.APIGatewayProxyResponse, error) {
	headers := map[string]string{
		"Content-Type":   "text",
		X_Slack_No_Retry: "1",
	}

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    headers,
		Body:       body.(string),
	}, nil
}

func ParseApigatewayParams(req events.APIGatewayProxyRequest) map[string]interface{} {
	paramsMap := make(map[string]interface{})
	// body must be in json format
	json.Unmarshal([]byte(req.Body), &paramsMap)
	queryMap := req.QueryStringParameters
	pathMap := req.PathParameters

	_maps := []map[string]string{queryMap, pathMap}
	for _, m := range _maps {
		for key, v := range m {
			paramsMap[key] = v
		}
	}
	return paramsMap
}

func IsRetryRequest(header http.Header) bool {
	return header[X_Slack_Retry_Num] != nil
}

package lambdaUtil

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
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
		"Content-Type": "application/json",
	}

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    headers,
		Body:       body.(string),
	}, nil
}

func BuildApigatewayJSONResponse(body interface{}, statusCode int, err error, header map[string]string) (events.APIGatewayProxyResponse, error) {
	r := StandardResponse{
		StatusCode: statusCode,
		Body:       body,
	}
	if err != nil {
		r.Message = err.Error()
	}
	bodystr, _ := json.Marshal(r)
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    header,
		Body:       string(bodystr),
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

func BuildHeader(allow_origin string) map[string]string {
	return map[string]string{
		"Access-Control-Allow-Headers": "Content-Type",
		"Access-Control-Allow-Origin":  allow_origin,
		"Access-Control-Allow-Methods": "OPTIONS,GET,POST",
	}
}

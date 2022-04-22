package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/mytn1992/ebd-carpark-streaming-api/pkg/cmd/random"
	"github.com/mytn1992/ebd-carpark-streaming-api/pkg/cmd/random_elastic"
	"github.com/mytn1992/ebd-carpark-streaming-api/pkg/cmd/streamingapi"
	"github.com/mytn1992/ebd-carpark-streaming-api/pkg/common/util"
)

// BodyRequest is our self-made struct to process JSON request from Client
type BodyRequest struct {
	Input string `json:"input"`
}

// BodyResponse is our self-made struct to build response for Client
type BodyResponse struct {
	ResponseName string `json:"name"`
}

type UserInput struct {
	Input    string `json:"input"`
	Distance int    `json:"distance"`
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

	result, statusCode, err := random.Run()
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
		// apiUrl := "https://8ucr04gpxf.execute-api.ap-southeast-1.amazonaws.com/dev/event"
		buildings := []random_elastic.Buildings{}
		err := util.OpenJSONFile("./data/buildingsShort.json", &buildings)
		if err != nil {
			log.Fatalf("Error while opening buildingsShort.json file: %v", err)
		}
		distanceChoice := []int{500, 700, 900}
		totalBuildings := len(buildings)
		for {
			no := rand.Intn(totalBuildings - 1)
			distance := distanceChoice[rand.Intn(len(distanceChoice))]

			// random local invoke
			buildings[no].Distance = int64(distance)
			buildings[no].Input = buildings[no].Postal
			fmt.Println(buildings[no])
			response, _, err := streamingapi.Run(buildings[no])
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(response)

			// random api invoke
			// userQuery := UserInput{
			// 	Input:    buildings[no].Postal,
			// 	Distance: distance,
			// }
			// jsonUserQuery, _ := json.Marshal(userQuery)
			// fmt.Println(string(jsonUserQuery))
			// err = util.SendPostRequest(apiUrl, jsonUserQuery)
			// if err != nil {
			// 	log.Fatal(err)
			// }

			sleepTime := rand.Intn(5)
			fmt.Printf("sleeping %v second\n", sleepTime)
			time.Sleep(time.Duration(sleepTime) * time.Second)
		}
	}
}

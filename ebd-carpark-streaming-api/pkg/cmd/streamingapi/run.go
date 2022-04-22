package streamingapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/fatih/structs"
	"github.com/mytn1992/ebd-carpark-streaming-api/pkg/cmd/random_elastic"
	"github.com/mytn1992/ebd-carpark-streaming-api/pkg/common/esw"
	"github.com/mytn1992/ebd-carpark-streaming-api/pkg/common/util"

	"github.com/olivere/elastic/v7"
	log "github.com/sirupsen/logrus"
)

const (
	// 	ES_Username = ""
	// 	ES_Password = ""
	// 	ES_Host     = ""
	//  Index       = ""
	ES_Username      = ""
	ES_Password      = ""
	ES_Host          = ""
	Index            = ""
	bootstrapServers = ""
	ccRestProxy      = ""
	ccloudAPIKey     = ""
	ccloudAPISecret  = ""
)

func Run(body random_elastic.Buildings) ([]map[string]interface{}, int, error) {
	// topic := "USER_QUERY"
	returnResult := []map[string]interface{}{}
	esWrapper, err := esw.NewWrapper(esw.Config{
		Username: ES_Username,
		Password: ES_Password,
		Host:     ES_Host,
	})
	if err != nil {
		log.Errorf("error in getting elastic client - %v", err)
		return nil, http.StatusInternalServerError, err
	}

	tempPostal, err := getOneMapResponse(body.Input)
	if err != nil {
		log.Fatal(err)
	}
	body.Latitude = tempPostal.Latitude
	body.Longitude = tempPostal.Longitude
	body.Postal = tempPostal.Postal
	body.Update_datetime = time.Now().UTC()
	geoPoint, err := elastic.GeoPointFromString(body.Latitude + "," + body.Longitude)
	if err != nil {
		log.Fatal(err)
	}
	body.Location = geoPoint
	query, geoSort := buildQueryForNearestCarpark(body)
	result, err := esw.QueryDataForAPI(*esWrapper, Index, query, true, []string{"", ""}, geoSort, 3)
	if err != nil {
		log.Errorf("error queryDataForAPI: %v", err)
	}

	queryResult := []PredictionResult{}
	for _, v := range result {
		query, timeSort := buildQueryForPredictionData(v["CARPARK_NO"].(string))
		predictionResponse, err := esw.QueryDataForAPI(*esWrapper, "hdb_cpk_processed", query, true, []string{"", ""}, timeSort, 10)
		if err != nil {
			log.Errorf("error queryDataForAPI: %v", err)
		}
		fmt.Println(predictionResponse)
		tempResult := PredictionResult{}
		tempResult.Car_park_no = v["CARPARK_NO"].(string)
		tempResult.Address = v["ADDRESS"].(string)
		tempResult.Latitude = v["LATITUDE"].(string)
		tempResult.Longitude = v["LONGITUDE"].(string)
		geoPointTemp, _ := elastic.GeoPointFromString(tempResult.Latitude + "," + tempResult.Longitude)
		tempResult.Location = geoPointTemp
		tempResult.Distance = fmt.Sprintf("%.2f", v["sortValue"])
		tempResult.Current_Lots_Available = fmt.Sprintf("%v", v["LOTS_AVAILABLE"])
		temptime := v["UPDATE_DATETIME"].(string)[:16]
		myDate, err := time.Parse("2006-01-02T15:04", temptime)
		if err != nil {
			panic(err)
		}
		tempResult.Time = myDate.String()[:16]
		for _, vv := range predictionResponse {
			tempResult.PredictionPointList = append(tempResult.PredictionPointList,
				PredictionPoint{
					Time:  vv["update_datetime"].(string)[:16],
					Value: vv["yhat"].(string),
				},
			)
		}
		queryResult = append(queryResult, tempResult)
	}
	for _, v := range queryResult {
		fmt.Printf("\nCarpark No: %v, Address: %v, Distance:%vm\nUpdate Time:%v, Current Lots Available: %v\n", v.Car_park_no, v.Address, v.Distance, v.Time, v.Current_Lots_Available)
		for _, vv := range v.PredictionPointList {
			fmt.Printf("Time: %v, Estimated Lots Available: %v\n", vv.Time, vv.Value)
		}
		fmt.Println("")
	}
	fmt.Println(queryResult)

	// kafka rest proxy invoke
	kafkaBody := KafkaStreamBody{
		ResultList: queryResult,
		Userinput:  body,
	}
	returnResult = append(returnResult, structs.Map(kafkaBody))

	tempPayload := KafkaPayload{}
	tempPayload.Records = append(tempPayload.Records,
		Payload{
			Key:   fmt.Sprintf("%v-%v", time.Now().Format("2006-01-02 15:04:05"), body.Input),
			Value: kafkaBody,
		},
	)
	jsonPayload, _ := json.Marshal(tempPayload)

	fmt.Println(string(jsonPayload))
	// kafka rest
	response, err := util.SendKafkaPostRequest(ccRestProxy, jsonPayload)
	if err != nil {
		log.Fatalf("error sending POST request to rest url: %v", err)
	}
	fmt.Println(string(response))

	return returnResult, http.StatusOK, nil
}

func buildQueryForNearestCarpark(body random_elastic.Buildings) (*elastic.BoolQuery, *elastic.GeoDistanceSort) {
	timeRange := elastic.NewRangeQuery("API_TIMESTAMP").Gte(("now-30m"))
	lotsRange := elastic.NewRangeQuery("LOTS_AVAILABLE").Gt(0)
	templatlong, err := elastic.GeoPointFromString(body.Latitude + "," + body.Longitude)
	if err != nil {
		log.Fatal(err)
	}
	geoQuery := elastic.NewGeoDistanceQuery("LOCATION").Distance(fmt.Sprintf("%vm", body.Distance)).Point(templatlong.Lat, templatlong.Lon)
	geoSort := elastic.NewGeoDistanceSort("LOCATION").Asc().Points(templatlong)
	return elastic.NewBoolQuery().Must(timeRange, lotsRange, geoQuery), geoSort
}

func buildQueryForPredictionData(car_park_no string) (*elastic.BoolQuery, *elastic.FieldSort) {
	timeRange := elastic.NewRangeQuery("update_datetime").Gte(("now+8h")).Lte("now+10h")
	termQuery := elastic.NewTermQuery("CARPARK_NO.keyword", car_park_no)
	fieldSort := elastic.NewFieldSort("update_datetime")
	return elastic.NewBoolQuery().Must(timeRange, termQuery), fieldSort
}

func getOneMapResponse(body string) (Result, error) {

	onemapResponse := &OnemapResponse{}
	q := url.Values{}
	q.Add("searchVal", body)
	q.Add("returnGeom", "Y")
	q.Add("getAddrDetails", "Y")
	resBody, err := util.SendGETRequest("https://developers.onemap.sg/commonapi/search?" + q.Encode())
	if err != nil {
		log.Errorf("error SendGETRequest: %v", err)
		return Result{}, err
	}
	err = json.Unmarshal(resBody, onemapResponse)
	if err != nil {
		log.Errorf("error Unmarshal: %v", err)
		return Result{}, err
	}
	if len(onemapResponse.Results) != 0 {
		fmt.Println("Postal:", onemapResponse.Results[0].Postal)
		fmt.Println("Lattitude:", onemapResponse.Results[0].Latitude)
		fmt.Println("Longitude:", onemapResponse.Results[0].Longitude)
	}
	return onemapResponse.Results[0], nil
}

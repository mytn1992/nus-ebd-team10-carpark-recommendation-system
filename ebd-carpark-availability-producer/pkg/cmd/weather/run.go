package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/mytn1992/ebd-carpark-availability-producer/pkg/common/esw"
	"github.com/mytn1992/ebd-carpark-availability-producer/pkg/common/util"
	"github.com/olivere/elastic"
)

const (
	bootstrapServers = ""
	ccRestProxy      = ""
	ccloudAPIKey     = ""
	ccloudAPISecret  = ""
	ES_Username      = ""
	ES_Password      = ""
	ES_Host          = ""
)

func Run() (string, int, error) {
	httpPostUrl := "https://api.data.gov.sg/v1/environment/rainfall?"
	start := time.Now()
	q := url.Values{}
	q.Add("date_time", time.Now().Format("2006-01-02T15:04:05"))
	response, err := util.SendGetRequest(httpPostUrl+q.Encode(), nil)
	if err != nil {
		log.Fatalf("error sending POST request to response url: %v", err)
	}

	var result Results
	err = json.Unmarshal([]byte(response), &result)
	if err != nil {
		log.Fatalf("error unmarshalling result: %v", err)
	}
	fmt.Println(string(response))
	fmt.Println(len(result.Metadata.Stations))
	fmt.Println(len(result.Items[0].Readings))

	for k, v := range result.Metadata.Stations {
		for _, vv := range result.Items[0].Readings {
			if vv.StationID == v.ID {
				result.Metadata.Stations[k].Value = vv.Value
				result.Metadata.Stations[k].UpdateDatetime = time.Now().UTC()
				result.Metadata.Stations[k].Device_location = elastic.GeoPointFromLatLon(v.Location.Latitude, v.Location.Longitude)
				result.Metadata.Stations[k].Location_string = fmt.Sprintf("%.6f,%.6f", v.Location.Latitude, v.Location.Longitude)
			}
		}
	}

	for _, v := range result.Metadata.Stations {
		tempPayload := KafkaPayload{}
		tempPayload.Records = append(tempPayload.Records,
			Payload{
				Key:   fmt.Sprintf("%v-%v", time.Now().Format("2006-01-02T15:04"), v.Name),
				Value: v,
			},
		)
		tempbody, _ := json.Marshal(tempPayload)
		fmt.Println(string(tempbody))
		response, err := util.SendKafkaPostRequest(ccRestProxy, tempbody)
		if err != nil {
			log.Fatalf("error sending POST request to response url: %v", err)
		}
		fmt.Println(string(response))
	}

	fmt.Println("done sending to kafka")
	duration := time.Since(start)
	fmt.Println("total duration:", duration)
	return "", http.StatusOK, nil

}

func ExportToElasticsearch(station []Station) error {
	esWrapper, err := esw.NewWrapper(esw.Config{
		Username: ES_Username,
		Password: ES_Password,
		Host:     ES_Host,
	})
	if err != nil {
		log.Errorf("error in getting elastic client - %v", err)
		return err
	}
	index := strings.Replace("rainfall-{DATE}", "{DATE}", time.Now().Format("2006-01-02"), -1)
	bulkRequest := esWrapper.ESClient.Bulk()

	for _, data := range station {
		// limit docID string length as it has a size limit of 512 bytes
		docID := fmt.Sprintf("%v-%v-%v-%v", time.Now().Format("2006-01-02T15:04:05"), data.Name, data.DeviceID, data.ID)
		indexReq := elastic.NewBulkIndexRequest().Index(index).Id(docID).Doc(data)
		bulkRequest = bulkRequest.Add(indexReq)
	}
	bulkResponse, err := bulkRequest.Do(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(bulkResponse)
	fmt.Println(err)
	if bulkResponse == nil {
		log.Errorf("expected bulkResponse to be != nil; got nil")
	}
	log.Info("done exporting to elastic")
	return nil
}

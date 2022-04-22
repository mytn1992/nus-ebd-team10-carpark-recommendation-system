package carparkinformation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/olivere/elastic/v7"
	log "github.com/sirupsen/logrus"

	"github.com/mytn1992/ebd-carpark-availability-producer/pkg/common/esw"
	"github.com/mytn1992/ebd-carpark-availability-producer/pkg/common/util"
)

type Results struct {
	Items []Item `json:"items"`
}

type Item struct {
	Timestamp   string         `json:"timestamp"`
	CarparkData []CarparkDatum `json:"carpark_data"`
}

type CarparkDatum struct {
	CarparkInfo    []CarparkInfo `json:"carpark_info"`
	CarparkNumber  string        `json:"carpark_number"`
	UpdateDatetime string        `json:"update_datetime"`
}

type CarparkInfo struct {
	TotalLots     string `json:"total_lots"`
	LotType       string `json:"lot_type"`
	LotsAvailable string `json:"lots_available"`
}

type OnemapResponse struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func Run() (string, int, error) {
	// util.LogSystemUsage(1 * time.Second)
	start := time.Now()
	httpPostUrl := "https://data.gov.sg/api/action/datastore_search?resource_id=139a3035-e624-4f56-b63f-89ae28d4ae4c&offset=%v"
	converterUrl := "https://developers.onemap.sg/commonapi/convert/3414to4326?"

	masterCarparkData := map[string]Record{}
	masterCarparkDataES := map[string]ESRecord{}
	i := 0
	total := 100000

	wg := sync.WaitGroup{}
	guardC := make(chan int, 10)
	mutex := &sync.Mutex{}
	for i < total {
		wg.Add(1)
		guardC <- 1
		go func(i int) {
			response, err := util.SendGetRequest(fmt.Sprintf(httpPostUrl, i), nil)
			if err != nil {
				log.Fatalf("error sending POST request to response url: %v", err)
			}
			var result CarparkInformation
			err = json.Unmarshal([]byte(response), &result)
			if err != nil {
				log.Fatalf("error unmarshalling result: %v", err)
			}

			for _, v := range result.Result.Records {

				onemapResponse := &OnemapResponse{}
				q := url.Values{}
				q.Add("X", v.XCoord)
				q.Add("Y", v.YCoord)
				resBody, err := util.SendGetRequest(converterUrl+q.Encode(), nil)
				if err != nil {
					log.Errorf("error SendGETRequest: %v", err)
				}
				err = json.Unmarshal(resBody, onemapResponse)
				if err != nil {
					log.Errorf("error Unmarshal: %v", err)
				}
				v.Latitude = fmt.Sprintf("%.6f", onemapResponse.Latitude)
				v.Longitude = fmt.Sprintf("%.6f", onemapResponse.Longitude)
				v.Location = elastic.GeoPointFromLatLon(onemapResponse.Latitude, onemapResponse.Longitude)
				mutex.Lock()
				masterCarparkData[strings.ToLower(v.CarParkNo)] = v
				mutex.Unlock()
			}
			total = int(result.Result.Total)
			fmt.Printf("i:%v-masterdata:%v-apitotal:%v-totalinloop:%v-workernum:%v\n", i, len(masterCarparkData), total, len(result.Result.Records), len(guardC))

			wg.Done()
			<-guardC
		}(i)
		i += 100
	}
	wg.Wait()

	httpPostUrl2 := "https://api.data.gov.sg/v1/transport/carpark-availability"
	carparkAvailability, err := util.SendGetRequest(httpPostUrl2, nil)
	if err != nil {
		log.Fatalf("error sending POST request to response url: %v", err)
	}

	var resultAvail Results
	err = json.Unmarshal([]byte(carparkAvailability), &resultAvail)
	if err != nil {
		log.Fatalf("error unmarshalling result: %v", err)
	}
	count := 0
	for _, carpark := range resultAvail.Items[0].CarparkData {
		if v, ok := masterCarparkData[strings.ToLower(carpark.CarparkNumber)]; ok { // myTime, err := time.Parse("2006-01-02T15:04:05", carpark.UpdateDatetime)

			tempRec := ESRecord{}
			tempRec.LotsAvailable, _ = strconv.Atoi(carpark.CarparkInfo[0].LotsAvailable)
			tempRec.TotalLots, _ = strconv.Atoi(carpark.CarparkInfo[0].TotalLots)
			tempRec.LotType = carpark.CarparkInfo[0].LotType
			tempRec.UpdateDatetime = time.Now().UTC()
			tempRec.ApiTimestamp = time.Now().UTC()
			tempRec.CarParkNo = v.CarParkNo
			tempRec.Address = v.Address
			tempRec.Location = v.Location
			tempRec.CarParkType = v.CarParkType
			tempRec.TypeOfParkingSystem = v.TypeOfParkingSystem
			tempRec.Latitude = v.Latitude
			tempRec.Longitude = v.Longitude
			tempgeo, _ := elastic.GeoPointFromString(v.Latitude + "," + v.Longitude)
			tempRec.Location = tempgeo

			masterCarparkDataES[strings.ToLower(carpark.CarparkNumber)] = tempRec
			count++
		}
	}
	finalRecords := []interface{}{}
	exportResult := []ESRecord{}

	for _, v := range masterCarparkDataES {
		finalRecords = append(finalRecords, v)
		exportResult = append(exportResult, v)
	}
	fmt.Println("total carpark availability data:", len(resultAvail.Items[0].CarparkData))
	fmt.Println("total carpark number:", len(masterCarparkData))
	fmt.Println("total merged:", count)

	err = ExportToElasticsearch(exportResult)
	if err != nil {
		log.Fatal(err)
	}

	// Code to measure
	duration := time.Since(start)

	// Formatted string, such as "2h3m0.5s" or "4.503μs"
	fmt.Println("total duration:", duration)
	return "", http.StatusOK, nil
}

func ExportToElasticsearch(carparkRecord []ESRecord) error {
	esWrapper, err := esw.NewWrapper(esw.Config{
		Username: "",
		Password: "",
		Host:     "",
	})
	if err != nil {
		log.Errorf("error in getting elastic client - %v", err)
		return err
	}
	index := strings.Replace("hdb_cpkall", "{DATE}", time.Now().Format("2006-01-02"), -1)
	bulkRequest := esWrapper.ESClient.Bulk()

	for _, data := range carparkRecord {
		// limit docID string length as it has a size limit of 512 bytes
		docID := fmt.Sprintf("%v-%v-%v-%v", time.Now().Format("2006-01-02T15:0"), data.CarParkNo, data.XCoord, data.YCoord)
		indexReq := elastic.NewBulkIndexRequest().Index(index).Id(docID).Doc(data)
		bulkRequest = bulkRequest.Add(indexReq)
	}
	bulkResponse, err := bulkRequest.Do(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	if bulkResponse == nil {
		log.Errorf("expected bulkResponse to be != nil; got nil")
	}
	log.Info(bulkResponse)
	log.Info("done exporting to elastic")
	return nil
}

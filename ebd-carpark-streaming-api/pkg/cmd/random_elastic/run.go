package random_elastic

import (
	"context"
	"fmt"
	"strconv"

	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/olivere/elastic"
	log "github.com/sirupsen/logrus"

	"github.com/mytn1992/ebd-carpark-streaming-api/pkg/common/esw"
	"github.com/mytn1992/ebd-carpark-streaming-api/pkg/common/util"
)

type Buildings struct {
	Address         string      `json:"ADDRESS"`
	Building        string      `json:"BUILDING"`
	Latitude        string      `json:"LATITUDE"`
	Longitude       string      `json:"LONGITUDE"`
	Location        interface{} `json:"location"`
	Update_datetime time.Time   `json:"update_datetime"`
	Postal          string      `json:"POSTAL"`
	Input           string      `json:"input"`
	Distance        int64       `json:"distance"`
}

func Run() (string, int, error) {
	//kafka producer

	buildings := []Buildings{}
	err := util.OpenJSONFile("./data/buildingsShort.json", &buildings)
	if err != nil {
		log.Fatalf("Error while opening buildingsShort.json file: %v", err)
	}
	totalBuildings := len(buildings)
	for {
		no := rand.Intn(totalBuildings - 1)
		fmt.Println(buildings[no])
		tempLocation := buildings[no]
		tempLat := 0.00
		tempLong := 0.00
		if lat, err := strconv.ParseFloat(tempLocation.Latitude, 64); err == nil {
			tempLat = lat
		}
		if long, err := strconv.ParseFloat(tempLocation.Longitude, 64); err == nil {
			tempLong = long
		}
		tempLocation.Update_datetime = time.Now().UTC()
		tempLocation.Location = elastic.GeoPointFromLatLon(tempLat, tempLong)
		ExportToElasticsearch([]Buildings{tempLocation})

		sleepTime := rand.Intn(5)
		fmt.Printf("sleeping %v second\n", sleepTime)
		time.Sleep(time.Duration(sleepTime) * time.Second)
	}

	return "", http.StatusOK, nil

}

func ExportToElasticsearch(carparkRecord []Buildings) error {
	esWrapper, err := esw.NewWrapper(esw.Config{
		Username: "",
		Password: "",
		Host:     "",
	})
	if err != nil {
		log.Errorf("error in getting elastic client - %v", err)
		return err
	}
	index := strings.Replace("realtime-userinput-{DATE}", "{DATE}", time.Now().Format("2006-01-02"), -1)
	bulkRequest := esWrapper.ESClient.Bulk()

	for _, data := range carparkRecord {
		// limit docID string length as it has a size limit of 512 bytes
		docID := fmt.Sprintf("%v-%v-", time.Now().Format("2006-01-02T15:04:05"), data.Postal)
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
	// log.Info(bulkResponse)
	log.Info("done exporting to elastic")
	return nil
}

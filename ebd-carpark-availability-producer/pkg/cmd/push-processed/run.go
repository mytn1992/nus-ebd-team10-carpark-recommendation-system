package push_processed

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/olivere/elastic/v7"
	log "github.com/sirupsen/logrus"

	"github.com/mytn1992/ebd-carpark-availability-producer/pkg/common/esw"
	"github.com/mytn1992/ebd-carpark-availability-producer/pkg/common/util"
)

const (
	ES_Username = ""
	ES_Password = ""
	ES_Host     = ""
	Index       = ""
)

type Results struct {
	Ds              string      `json:"ds" csv:"ds"`
	Car_park_no     string      `json:"car_park_no" csv:"car_park_no"`
	CARPARK_NO      string      `json:"CARPARK_NO" csv:"CARPARK_NO"`
	Y               string      `json:"y" csv:"y"`
	Y_int           interface{} `json:"y_int" csv:"y_int"`
	Yhat            string      `json:"yhat" csv:"yhat"`
	Yhat_int        int         `json:"yhat_int" csv:"yhat_int"`
	Update_datetime time.Time   `json:"update_datetime" csv:"update_datetime"`
	Location        interface{} `json:"LOCATION" csv:"LOCATION"`
	Latitude        string      `json:"latitude" csv:"latitude"`
	Longitude       string      `json:"longitude" csv:"longitude"`
}

func Run() (string, int, error) {
	util.LogSystemUsage(1 * time.Second)
	//kafka producer
	var files []string

	root := "/Users/matthew.yap/Documents/GoPratice/pyspark-examples/testdata_all_location_total_lots"
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if path[len(path)-4:] == ".csv" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		fmt.Println(file)
	}
	for _, v := range files[1:] {
		results := []Results{}
		err = util.OpenCSVFile(v, &results)
		if err != nil {
			log.Fatalf("Error while opening buildingsShort.json file: %v", err)
		}
		for k, v := range results {
			myDate, err := time.Parse("2006-01-02T15:04", v.Ds[:16])
			if err != nil {
				panic(err)
			}
			results[k].Update_datetime = myDate
			results[k].Y_int, _ = strconv.Atoi(v.Y)
			results[k].Yhat_int, _ = strconv.Atoi(v.Yhat)
			results[k].Latitude = v.Latitude
			results[k].Longitude = v.Longitude
			results[k].Location, _ = elastic.GeoPointFromString(v.Latitude + "," + v.Longitude)
			results[k].CARPARK_NO = v.Car_park_no
			if v.Y == "" {
				results[k].Y_int = nil
			}

		}
		ExportToElasticsearch(results)
	}

	return "", http.StatusOK, nil

}

func ExportToElasticsearch(carparkRecord []Results) error {
	esWrapper, err := esw.NewWrapper(esw.Config{
		Username: ES_Username,
		Password: ES_Password,
		Host:     ES_Host,
	})
	if err != nil {
		log.Errorf("error in getting elastic client - %v", err)
		return err
	}
	index := strings.Replace(Index+"-{DATE}", "{DATE}", time.Now().Format("2006-01-02"), -1)
	bulkRequest := esWrapper.ESClient.Bulk()

	for _, data := range carparkRecord {
		// limit docID string length as it has a size limit of 512 bytes
		docID := fmt.Sprintf("%v-%v-", data.Ds, data.Car_park_no)
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
	log.Info("done exporting to elastic")
	return nil
}

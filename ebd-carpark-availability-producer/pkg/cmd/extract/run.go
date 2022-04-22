package extract

import (
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	log "github.com/sirupsen/logrus"

	"github.com/mytn1992/ebd-carpark-availability-producer/pkg/common/esw"
	"github.com/mytn1992/ebd-carpark-availability-producer/pkg/common/s3w"
	"github.com/mytn1992/ebd-carpark-availability-producer/pkg/common/util"

	"github.com/olivere/elastic/v7"
)

const (
	ES_Username = ""
	ES_Password = ""
	ES_Host     = ""
	Index       = ""
)

type SampleModel struct {
	Update_datetime string `mapstructure:"update_datetime" csv:"ds" json:"update_datetime"`
	Lots_available  string `mapstructure:"lots_available" csv:"y" json:"lots_available"`
	Car_park_no     string `mapstructure:"car_park_no" csv:"car_park_no" json:"car_park_no"`
	Latitude        string `mapstructure:"latitude" csv:"latitude" json:"latitude"`
	Longitude       string `mapstructure:"longitude" csv:"longitude" json:"longitude"`
	Total_lots      string `mapstructure:"total_lots" csv:"total_lots" json:"total_lots"`
}

func Run() (string, int, error) {
	util.LogSystemUsage(1 * time.Second)
	esWrapper, err := esw.NewWrapper(esw.Config{
		Username: ES_Username,
		Password: ES_Password,
		Host:     ES_Host,
	})
	if err != nil {
		log.Errorf("error init wrapper: %v", err)
	}
	query := buildQueryForCarparkIndex("")
	result, err := esw.QueryScrollDataForAPI(*esWrapper, Index, query, false, []string{"UPDATE_DATETIME", "LOTS_AVAILABLE", "TOTAL_LOTS", "CARPARK_NO", "LATITUDE", "LONGITUDE", "API_TIMESTAMP"}, "UPDATE_DATETIME")
	if err != nil {
		log.Errorf("error queryDataForAPI: %v", err)
	}

	toSave := []interface{}{}
	for _, v := range result {
		temptime := v["API_TIMESTAMP"].(string)[:16]
		myDate, err := time.Parse("2006-01-02T15:04", temptime)
		if err != nil {
			panic(err)
		}

		dataTemp := SampleModel{
			Update_datetime: myDate.String()[:16],
			Lots_available:  v["LOTS_AVAILABLE"].(string),
			Car_park_no:     v["CARPARK_NO"].(string),
			Latitude:        util.Trim(v["LATITUDE"].(string), ".", 6),
			Longitude:       util.Trim(v["LONGITUDE"].(string), ".", 6),
			Total_lots:      v["TOTAL_LOTS"].(string),
		}
		toSave = append(toSave, dataTemp)
	}

	carparkData, err := util.WriteToCSV("/tmp/carpark_data_all_location.csv", toSave)
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("Exported to %v", *carparkData)

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("ap-southeast-1"),
	})
	if err != nil {
		log.Errorf("can't init session %v", err)
	}
	s3wrapper := s3w.NewWrapper(sess)
	err = s3wrapper.PutObject("ebd-demo", "input/"+GenFileName(), "/tmp/carpark_data_all_location.csv")
	if err != nil {
		log.Fatal(err)
	}
	log.Info("Successfully uploaded files to s3")
	return "", http.StatusOK, nil
}

func buildQueryForCarparkIndex(carpark_no string) *elastic.BoolQuery {
	dateQuery := elastic.NewRangeQuery("UPDATE_DATETIME").Gte(("now-3d"))
	return elastic.NewBoolQuery().Must(dateQuery)
}

func GenFileName() string {
	return fmt.Sprintf("carpark-all-location-%v.csv", time.Now().Format("20060102T15"))
}

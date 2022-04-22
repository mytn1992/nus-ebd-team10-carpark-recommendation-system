package query

import (
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/mytn1992/ebd-carpark-availability-producer/pkg/common/esw"
	"github.com/mytn1992/ebd-carpark-availability-producer/pkg/common/util"

	"github.com/olivere/elastic/v7"
)

type SampleModel struct {
	Update_datetime string  `mapstructure:"update_datetime" csv:"ds" json:"update_datetime"`
	Lots_available  float64 `mapstructure:"lots_available" csv:"y" json:"lots_available"`
	Total_lots      string  `mapstructure:"total_lots" csv:"total_lots" json:"total_lots"`
	Car_park_no     string  `mapstructure:"car_park_no" csv:"car_park_no" json:"car_park_no"`
	Latitude        string  `mapstructure:"latitude" csv:"latitude" json:"latitude"`
	Longitude       string  `mapstructure:"longitude" csv:"longitude" json:"longitude"`
}

func Run() (string, int, error) {
	util.LogSystemUsage(1 * time.Second)
	carpark_no := ""
	index := "carparkinformation"
	esWrapper, err := esw.NewWrapper(esw.Config{
		Username: "",
		Password: "",
		Host:     "",
	})
	if err != nil {
		log.Errorf("error init wrapper: %v", err)
	}
	query := buildQueryForCarparkIndex(carpark_no)
	result, err := esw.QueryScrollDataForAPI(*esWrapper, index, query, false, []string{"update_datetime", "lots_available", "total_lots", "car_park_no", "latitude", "longitude"}, "update_datetime")
	if err != nil {
		log.Errorf("error queryDataForAPI: %v", err)
	}
	fmt.Println(result)
	loc, err := time.LoadLocation("Singapore")
	if err != nil {
		panic(err)
	}

	toSave := []interface{}{}
	for _, v := range result {
		temptime := v["update_datetime"].(string)[:16]
		myDate, err := time.Parse("2006-01-02T15:04", temptime)
		if err != nil {
			panic(err)
		}
		tempdateTime := myDate.In(loc)

		dataTemp := SampleModel{
			Update_datetime: tempdateTime.String()[:16],
			Lots_available:  v["lots_available"].(float64),
			Total_lots:      fmt.Sprintf("%v", v["total_lots"]),
			Car_park_no:     v["car_park_no"].(string),
			Latitude:        v["latitude"].(string),
			Longitude:       v["longitude"].(string),
		}
		toSave = append(toSave, dataTemp)
	}

	carparkData, err := util.WriteToCSV("carpark_data_all_location_w_total_lots.csv", toSave)
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("Exported to %v", *carparkData)
	return "", http.StatusOK, nil
}

func buildQueryForCarparkIndex(carpark_no string) *elastic.BoolQuery {
	dateQuery := elastic.NewRangeQuery("update_datetime").Gte(("now-3d"))
	return elastic.NewBoolQuery().Must(dateQuery)
}

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/mytn1992/ebd-carpark-streaming-api/pkg/common/util"
)

type ReducedBuildings struct {
	Address   string `json:"ADDRESS"`
	Building  string `json:"BUILDING"`
	Latitude  string `json:"LATITUDE"`
	Longitude string `json:"LONGITUDE"`
	Postal    string `json:"POSTAL"`
}

func main() {
	buildings := []ReducedBuildings{}
	err := util.OpenJSONFile("./data/buildings.json", &buildings)
	if err != nil {
		log.Fatalf("Error while opening buildings.json file: %v", err)
	}

	buildingJson, _ := json.Marshal(buildings)
	err = ioutil.WriteFile("outputBuidling.json", buildingJson, 0644)
	if err != nil {
		log.Fatalf("Error while saving outputBuidling.json file: %v", err)
	}
	fmt.Println(len(buildings))
}

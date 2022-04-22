package random

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/mytn1992/ebd-carpark-streaming-api/pkg/common/util"
)

type Buildings struct {
	Address   string `json:"ADDRESS"`
	Building  string `json:"BUILDING"`
	Latitude  string `json:"LATITUDE"`
	Longitude string `json:"LONGITUDE"`
	Postal    string `json:"POSTAL"`
}

func Run() (string, int, error) {
	topic := "user-input"
	//kafka producer
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"client.id":         "1",
		"acks":              "all"})

	if err != nil {
		fmt.Printf("Failed to create producer: %s\n", err)
		os.Exit(1)
	}

	// Go-routine to handle message delivery reports and
	// possibly other event types (errors, stats, etc)
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Failed to deliver message: %v\n", ev.TopicPartition)
				} else {
					fmt.Printf("Successfully produced record to topic %s partition [%d] @ offset %v\n",
						*ev.TopicPartition.Topic, ev.TopicPartition.Partition, ev.TopicPartition.Offset)
				}
			}
		}
	}()

	buildings := []Buildings{}
	err = util.OpenJSONFile("./data/buildingsShort.json", &buildings)
	if err != nil {
		log.Fatalf("Error while opening buildingsShort.json file: %v", err)
	}
	totalBuildings := len(buildings)
	for {
		no := rand.Intn(totalBuildings - 1)
		fmt.Println(buildings[no])
		jDoc, err := json.Marshal(buildings[no])
		if err != nil {
			log.Fatalf("Unexpected error: %s", err)
		}
		//kafka produce
		p.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Key:            []byte(fmt.Sprintf("%v-%v", time.Now().Format("2006-01-02 15:04:05"), buildings[no].Postal)),
			Value:          jDoc,
		}, nil)

		sleepTime := rand.Intn(5)
		fmt.Printf("sleeping %v second\n", sleepTime)
		time.Sleep(time.Duration(sleepTime) * time.Second)
	}

	p.Flush(15 * 1000)
	p.Close()
	fmt.Println("done sending to kafka")
	return "", http.StatusOK, nil

}

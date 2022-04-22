package carparkavailability

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/mytn1992/ebd-carpark-availability-producer/pkg/common/util"
)

const (
	bootstrapServers = ""
	ccloudAPIKey     = ""
	ccloudAPISecret  = ""
)

func Run() (string, int, error) {
	httpPostUrl := "https://api.data.gov.sg/v1/transport/carpark-availability"
	topic := "HDB_CPK_AVAILABILITY"

	response, err := util.SendGetRequest(httpPostUrl, nil)
	if err != nil {
		log.Fatalf("error sending POST request to response url: %v", err)
	}

	var result Results
	err = json.Unmarshal([]byte(response), &result)
	if err != nil {
		log.Fatalf("error unmarshalling result: %v", err)
	}

	fmt.Println(len(result.Items[0].CarparkData))

	//kafka producer
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		"sasl.mechanisms":   "PLAIN",
		"security.protocol": "SASL_SSL",
		"sasl.username":     ccloudAPIKey,
		"sasl.password":     ccloudAPISecret})

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

	for _, r := range result.Items[0].CarparkData {
		jDoc, err := json.Marshal(r)
		if err != nil {
			log.Fatalf("Unexpected error: %s", err)
		}

		//kafka produce
		p.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Key:            []byte(fmt.Sprintf("%v-%v", time.Now().Format("2006-01-02 15:04"), r.CarparkNumber)),
			Value:          jDoc,
		}, nil)
	}

	p.Flush(15 * 1000)
	p.Close()
	fmt.Println("done sending to kafka")
	return "", http.StatusOK, nil

}

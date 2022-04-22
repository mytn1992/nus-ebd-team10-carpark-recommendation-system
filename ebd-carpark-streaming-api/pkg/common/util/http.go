package util

import (
	"bytes"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func SendPostRequest(endpoint string, body []byte) error {
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		log.Errorf("err while sending POST request: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	log.Infof("response Status: %v", resp.Status)
	log.Infof("response Headers: %v", resp.Header)
	resp_body, _ := ioutil.ReadAll(resp.Body)
	log.Infof("response Body: %v", string(resp_body))
	return nil
}

func SendGETRequest(endpoint string) ([]byte, error) {
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		log.Errorf("err while sending POST request: %v", err)
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	resp_body, _ := ioutil.ReadAll(resp.Body)
	// log.Infof("response Body: %v", string(resp_body))
	return resp_body, nil
}

func SendKafkaPostRequest(endpoint string, body []byte) (response []byte, err error) {
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		log.Fatalf("err while sending POST request: %v", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/vnd.kafka.json.v2+json")
	req.Header.Set("Accept", "application/vnd.kafka.v2+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	resp_body, _ := ioutil.ReadAll(resp.Body)
	return resp_body, nil
}

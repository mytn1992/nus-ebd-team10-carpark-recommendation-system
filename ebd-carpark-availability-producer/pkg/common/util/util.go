package util

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func SendGetRequest(endpoint string, body []byte) (response []byte, err error) {
	req, err := http.NewRequest("GET", endpoint, bytes.NewBuffer(body))
	if err != nil {
		log.Fatalf("err while sending GET request: %v", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	resp_body, _ := ioutil.ReadAll(resp.Body)
	return resp_body, nil
}

func SendKafkaPostRequest(endpoint string, body []byte) (response []byte, err error) {
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		log.Fatalf("err while sending POST request: %v", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/vnd.kafka.json.v2+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	resp_body, _ := ioutil.ReadAll(resp.Body)
	return resp_body, nil
}

func Trim(body string, char string, length int) string {
	index := strings.Index(body, char)
	return body[:index+length+1]
}

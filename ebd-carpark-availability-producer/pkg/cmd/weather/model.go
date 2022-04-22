package weather

import "time"

type Results struct {
	Metadata Metadata `json:"metadata"`
	Items    []Item   `json:"items"`
	APIInfo  APIInfo  `json:"api_info"`
}

type APIInfo struct {
	Status string `json:"status"`
}

type Item struct {
	Timestamp string    `json:"timestamp"`
	Readings  []Reading `json:"readings"`
}

type Reading struct {
	StationID string  `json:"station_id"`
	Value     float64 `json:"value"`
}

type Metadata struct {
	Stations    []Station `json:"stations"`
	ReadingType string    `json:"reading_type"`
	ReadingUnit string    `json:"reading_unit"`
}

type KafkaPayload struct {
	Records []Payload `json:"records"`
}
type Payload struct {
	Key   string  `json:"key"`
	Value Station `json:"value"`
}

type Station struct {
	ID              string      `json:"id"`
	DeviceID        string      `json:"device_id"`
	Name            string      `json:"name"`
	Location        Location    `json:"location"`
	Value           float64     `json:"value"`
	UpdateDatetime  time.Time   `json:"update_datetime"`
	Device_location interface{} `json:"device_location"`
	Location_string string      `json:"location_string"`
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

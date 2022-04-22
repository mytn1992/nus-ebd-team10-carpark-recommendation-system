package carparkavailability

type Results struct {
	Items []Item `json:"items"`
}

type Item struct {
	Timestamp   string         `json:"timestamp"`
	CarparkData []CarparkDatum `json:"carpark_data"`
}

type CarparkDatum struct {
	CarparkInfo    []CarparkInfo `json:"carpark_info"`
	CarparkNumber  string        `json:"carpark_number"`
	UpdateDatetime string        `json:"update_datetime"`
}

type CarparkInfo struct {
	TotalLots     string `json:"total_lots"`
	LotType       string `json:"lot_type"`
	LotsAvailable string `json:"lots_available"`
}

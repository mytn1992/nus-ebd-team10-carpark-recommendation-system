package carparkinformation

import (
	"time"
)

type CarparkInformation struct {
	Help    string `json:"help"`
	Success bool   `json:"success"`
	Result  Result `json:"result"`
}

type Result struct {
	ResourceID string   `json:"resource_id"`
	Fields     []Field  `json:"fields"`
	Records    []Record `json:"records"`
	Links      Links    `json:"_links"`
	Total      int64    `json:"total"`
}

type Field struct {
	Type Type   `json:"type"`
	ID   string `json:"id"`
}

type Links struct {
	Start string `json:"start"`
	Next  string `json:"next"`
}

type Record struct {
	ShortTermParking ShortTermParking `json:"short_term_parking"`
	CarParkType      CarParkType      `json:"car_park_type"`
	YCoord           string           `json:"y_coord"`
	XCoord           string           `json:"x_coord"`
	FreeParking      FreeParking      `json:"free_parking"`
	GantryHeight     string           `json:"gantry_height"`
	CarParkBasement  CarParkBasement  `json:"car_park_basement"`
	NightParking     NightParking     `json:"night_parking"`
	Address          string           `json:"address"`
	CarParkDecks     string           `json:"car_park_decks"`
	// ID                  int64               `json:"_id"`
	CarParkNo           string              `json:"car_park_no"`
	TypeOfParkingSystem TypeOfParkingSystem `json:"type_of_parking_system"`
	UpdateDatetime      time.Time           `json:"update_datetime"`
	TotalLots           int                 `json:"total_lots"`
	LotType             string              `json:"lot_type"`
	LotsAvailable       int                 `json:"lots_available"`
	Latitude            string              `json:"latitude"`
	Longitude           string              `json:"longitude"`
	Location            interface{}         `json:"location"`
}

// backup
type ESRecord struct {
	ShortTermParking ShortTermParking `json:"SHORT_TERM_PARKING"`
	CarParkType      CarParkType      `json:"CARPARK_TYPE"`
	YCoord           string           `json:"Y_COORD"`
	XCoord           string           `json:"X_COORD"`
	FreeParking      FreeParking      `json:"FREE_PARKING"`
	GantryHeight     string           `json:"GANTRY_HEIGHT"`
	CarParkBasement  CarParkBasement  `json:"CARPARK_BASEMENT"`
	NightParking     NightParking     `json:"NIGHT_PARKING"`
	Address          string           `json:"ADDRESS"`
	CarParkDecks     string           `json:"CARPARK_DECKS"`
	// ID                  int64               `json:"_id"`
	CarParkNo           string              `json:"CARPARK_NO"`
	TypeOfParkingSystem TypeOfParkingSystem `json:"TYPE_OF_PARKING_SYSTEM"`
	ApiTimestamp        time.Time           `json:"API_TIMESTAMP"`
	UpdateDatetime      time.Time           `json:"UPDATE_DATETIME"`
	TotalLots           int                 `json:"TOTAL_LOTS"`
	LotType             string              `json:"LOT_TYPE"`
	LotsAvailable       int                 `json:"LOTS_AVAILABLE"`
	Latitude            string              `json:"LATITUDE"`
	Longitude           string              `json:"LONGITUDE"`
	Location            interface{}         `json:"LOCATION"`
}

type Type string

const (
	Int4    Type = "int4"
	Numeric Type = "numeric"
	Text    Type = "text"
)

type CarParkBasement string

const (
	N CarParkBasement = "N"
	Y CarParkBasement = "Y"
)

type CarParkType string

const (
	BasementCarPark    CarParkType = "BASEMENT CAR PARK"
	MultiStoreyCarPark CarParkType = "MULTI-STOREY CAR PARK"
	SurfaceCarPark     CarParkType = "SURFACE CAR PARK"
)

type FreeParking string

const (
	FreeParkingNO    FreeParking = "NO"
	SunPhFr7Am1030Pm FreeParking = "SUN & PH FR 7AM-10.30PM"
)

type NightParking string

const (
	NightParkingNO NightParking = "NO"
	Yes            NightParking = "YES"
)

type ShortTermParking string

const (
	ShortTermParkingNO ShortTermParking = "NO"
	The7Am1030Pm       ShortTermParking = "7AM-10.30PM"
	The7Am7Pm          ShortTermParking = "7AM-7PM"
	WholeDay           ShortTermParking = "WHOLE DAY"
)

type TypeOfParkingSystem string

const (
	CouponParking     TypeOfParkingSystem = "COUPON PARKING"
	ElectronicParking TypeOfParkingSystem = "ELECTRONIC PARKING"
)

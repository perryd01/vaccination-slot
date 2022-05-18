package chaincode

import (
	"encoding/json"
	"log"
	"time"
)

type VaccinationType string

const (
	Alpha   VaccinationType = "alpha"
	Bravo   VaccinationType = "bravo"
	Charlie VaccinationType = "charlie"
	Delta   VaccinationType = "delta"
	Echo    VaccinationType = "echo"
)

func (vt *VaccinationType) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	*vt = VaccinationType(s)
	if err != nil {
		return err
	}
	return nil
}

func (vt *VaccinationType) MarshalJSON() ([]byte, error) {
	return []byte("\"" + string(*vt) + "\""), nil
}

var deadlines = map[VaccinationType]time.Duration{}

func init() {
	dAlpha, err := time.ParseDuration("720h")
	if err != nil {
		log.Fatal(err)
	}
	deadlines[Alpha] = dAlpha
	dBravo, err := time.ParseDuration("720h")
	if err != nil {
		log.Fatal(err)
	}
	deadlines[Bravo] = dBravo
	dCharlie, err := time.ParseDuration("720h")
	if err != nil {
		log.Fatal(err)
	}
	deadlines[Charlie] = dCharlie
	dDelta, err := time.ParseDuration("720h")
	if err != nil {
		log.Fatal(err)
	}
	deadlines[Delta] = dDelta
	dEcho, err := time.ParseDuration("720h")
	if err != nil {
		log.Fatal(err)
	}
	deadlines[Echo] = dEcho
}

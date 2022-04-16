package chaincode

import (
	"encoding/json"
	"errors"
)

type VaccinationType string

const (
	Alpha   = "alpha"
	Bravo   = "bravo"
	Charlie = "charlie"
	Delta   = "delta"
	Echo    = "echo"
)

func (vt *VaccinationType) FromString(s []byte) error {
	switch string(s) {
	case "alpha":
		*vt = Alpha
	case "bravo":
		*vt = Bravo
	case "charlie":
		*vt = Charlie
	case "delta":
		*vt = Delta
	case "echo":
		*vt = Echo
	default:
		return errors.New("vaccinationType from string failed")
	}
	return nil
}

func (vt *VaccinationType) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	err = vt.FromString(data)
	if err != nil {
		return err
	}
	return nil
}

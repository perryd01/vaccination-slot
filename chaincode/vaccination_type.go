package chaincode

import (
	"encoding/json"
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

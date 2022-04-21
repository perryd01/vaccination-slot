package config

import (
	_ "embed"
	"encoding/json"
	"log"
)

//go:embed network.json
var network []byte

var n *Network

func init() {
	ne := &Network{}
	err := json.Unmarshal(network, ne)
	if err != nil {
		log.Println(err)
	} else {
		n = ne
	}
}

type Network struct {
	Organizations []string `json:"organizations"`
	Channel       string   `json:"channel"`
	DoctorMspid   string   `json:"doctor_mspid"`
	PatientMspid  string   `json:"patient_mspid"`
}

func NetworkConfig() *Network {
	return n
}

package main

import (
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-contract-api-go/metadata"
	cc "github.com/perryd01/vaccination-slot/chaincode"
	"log"
)

func main() {
	contract := &cc.VaccinationContract{
		IdGenerator: &cc.TokenIdGenerator{},
	}
	contract.Info.Version = "1.0.0"
	contract.Info.Description = "VaccinationSlots chaincode"
	contract.Info.License = &metadata.LicenseMetadata{}
	contract.Info.License.Name = "MIT"
	contract.Info.License.URL = "https://github.com/perryd01/vaccination-slot/blob/main/LICENSE"
	contract.Info.Contact = &metadata.ContactMetadata{}
	contract.Info.Contact.Name = "perryd01"

	chaincode, err := contractapi.NewChaincode(contract)

	if err != nil {
		log.Fatalf("could not create chaincode from VaccinationContract. %s", err)
	}

	chaincode.Info.Title = "Vaccination slot"
	chaincode.Info.Version = contract.Info.Version

	err = chaincode.Start()

	if err != nil {
		log.Fatalf("failed to start chaincode. %s", err)
	}
}

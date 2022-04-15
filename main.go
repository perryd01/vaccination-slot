package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/perryd01/vaccination-slot/chaincode"
)

func main() {
	vsToken := new(chaincode.VSTokenContract)

	chaincode, err := contractapi.NewChaincode(vsToken)
	chaincode.Info.Title = "ERC-721 chaincode"
	chaincode.Info.Version = "0.0.1"

	if err != nil {
		log.Fatal("Could not create chaincode " + err.Error())
	}

	err = chaincode.Start()

	if err != nil {
		panic("Failed to start chaincode. " + err.Error())
	}
}

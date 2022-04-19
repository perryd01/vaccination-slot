package chaincode

import (
	"encoding/base64"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// VaccinationContract is a smart contract for managing vaccination slots.
// Implements ERC-721.
type VaccinationContract struct {
	contractapi.Contract
}

func (c *VaccinationContract) ClientAccountId(ctx contractapi.TransactionContextInterface) (string, error) {
	clientAccountId64, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", fmt.Errorf("failed to get client identity")
	}

	clientAccountBytes, err := base64.StdEncoding.DecodeString(clientAccountId64)
	if err != nil {
		return "", fmt.Errorf("failed to decode string clientAccountId64: %v", err)
	}
	return string(clientAccountBytes), nil
}

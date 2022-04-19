package chaincode

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func readVaccinationSlot(ctx contractapi.TransactionContextInterface, tokenId string) (*VaccinationSlot, error) {
	key, err := ctx.GetStub().CreateCompositeKey(vsPrefix, []string{tokenId})
	if err != nil {
		return nil, fmt.Errorf("failed to create CompositeKey %s: %v", tokenId, err)
	}

	vsBytes, err := ctx.GetStub().GetState(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get state %s: %v", key, err)
	}

	vs := &VaccinationSlot{}
	err = json.Unmarshal(vsBytes, vs)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal vsBytes: %v", err)
	}

	return vs, nil
}

func vaccinationSlotExists(ctx contractapi.TransactionContextInterface, tokenId string) (bool, error) {
	key, err := ctx.GetStub().CreateCompositeKey(vsPrefix, []string{tokenId})
	if err != nil {
		return false, fmt.Errorf("failed to create CompositeKey %s: %v", tokenId, err)
	}

	vsBytes, err := ctx.GetStub().GetState(key)
	if err != nil {
		return false, fmt.Errorf("failed to get state %s: %v", key, err)
	}

	return len(vsBytes) > 0, nil
}

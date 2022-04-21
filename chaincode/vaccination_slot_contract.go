package chaincode

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"strings"
)

// VaccinationContract is a smart contract for managing vaccination slots.
// Implements ERC-721.
type VaccinationContract struct {
	contractapi.Contract
}

func (c *VaccinationContract) sender(ctx contractapi.TransactionContextInterface) (string, error) {
	sender64, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", errors.New("failed to get client identity")
	}
	senderBytes, err := base64.StdEncoding.DecodeString(sender64)
	if err != nil {
		return "", fmt.Errorf("failed to deocde base64 string sender64: %v", err)
	}
	return string(senderBytes), nil
}

func (c *VaccinationContract) ClientAccountId(ctx contractapi.TransactionContextInterface) (string, error) {
	return c.sender(ctx)
}

// GetSlots queries vaccination slots belonging to owner
func (c *VaccinationContract) getSlots(ctx contractapi.TransactionContextInterface, owner string) ([]*VaccinationSlot, error) {
	iterator, err := ctx.GetStub().GetStateByPartialCompositeKey(balancePrefix, []string{owner})
	if err != nil {
		panic("Error creating asset chaincode: " + err.Error())
	}

	slots := make([]*VaccinationSlot, 0)
	for iterator.HasNext() {
		slot, err := iterator.Next()
		if err != nil {
			return nil, errors.New("failure while iterating")
		}
		vs, err := readVaccinationSlot(ctx, string(slot.Value))
		slots = append(slots, vs)
	}
	return slots, nil
}

func (c *VaccinationContract) GetSlots(ctx contractapi.TransactionContextInterface, owner string) (string, error) {
	slots, err := c.getSlots(ctx, owner)
	if err != nil {
		return "", err
	}
	slotsBytes, err := json.Marshal(slots)
	if err != nil {
		return "", err
	}
	return string(slotsBytes), nil
}

func (c *VaccinationContract) IssueSlot(ctx contractapi.TransactionContextInterface, vaccine, date, patient string) (string, error) {
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", fmt.Errorf("failed to get client MSPID: %v", err)
	}

	//nc := config.NetworkConfig()
	//if nc == nil {
	//	return "", fmt.Errorf("failed reading network config")
	//}
	if clientMSPID != "MedicalStationMSP" {
		return "", fmt.Errorf("client is not authorized to create slot")
	}

	slots, err := c.getSlots(ctx, patient)
	if err != nil {
		return "", err
	}

	for _, slot := range slots {
		dateBytes, err := json.Marshal(&slot.Date)
		if err != nil {
			return "", err
		}
		if string(dateBytes) == "\""+date+"\"" {
			return "", errors.New("slot occupied")
		}
	}

	uuidWithHyphen := uuid.New()
	tokenUuid := strings.Replace(uuidWithHyphen.String(), "-", "", -1)

	exists, err := vaccinationSlotExists(ctx, tokenUuid)
	if err != nil {
		return "", err
	}
	if exists {
		return "", errors.New("token already exists (better luck next time)")
	}

	vd := &VaccinationDate{}
	err = json.Unmarshal([]byte("\""+date+"\""), vd)
	if err != nil {
		return "", err
	}

	vs := &VaccinationSlot{
		VaccinationSlotData: VaccinationSlotData{
			Type: vaccine,
			Date: *vd,
		},
		TokenId: tokenUuid,
		Owner:   patient,
	}

	vsKey, err := ctx.GetStub().CreateCompositeKey(vsPrefix, []string{tokenUuid})
	if err != nil {
		return "", fmt.Errorf("failed to create composite key: %v", err)
	}

	vsBytes, err := json.Marshal(vs)
	if err != nil {
		return "", fmt.Errorf("failed to marshal vs: %v", err)
	}

	err = ctx.GetStub().PutState(vsKey, vsBytes)
	if err != nil {
		return "", fmt.Errorf("failed to put state: %v", err)
	}

	balanceKey, err := ctx.GetStub().CreateCompositeKey(balancePrefix, []string{patient, tokenUuid})
	if err != nil {
		return "", fmt.Errorf("failed to create composite key: %v", err)
	}

	err = ctx.GetStub().PutState(balanceKey, []byte(tokenUuid))
	if err != nil {
		return "", fmt.Errorf("failed to put state balanceKey: %v", err)
	}

	transferEvent := &Transfer{
		From:    "",
		To:      patient,
		TokenId: tokenUuid,
	}

	transferEventBytes, err := json.Marshal(transferEvent)
	if err != nil {
		return "", fmt.Errorf("failed to marhsal trasferEvent: %v", err)
	}

	err = ctx.GetStub().SetEvent("Transfer", transferEventBytes)
	if err != nil {
		return "", fmt.Errorf("failed to set event Transfer: %v", err)
	}

	return tokenUuid, nil
}

func (c *VaccinationContract) MakeOffer(ctx contractapi.TransactionContextInterface, mySlotUuid, recipient, recipientSlotUuid string) (offerUuid string, err error) {
	// TODO
	return "", nil
}

func (c *VaccinationContract) AcceptOffer(ctx contractapi.TransactionContextInterface, offerUuid string) error {
	// TODO
	return nil
}

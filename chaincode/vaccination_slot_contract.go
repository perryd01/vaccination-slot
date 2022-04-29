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

	vt := new(VaccinationType)
	err = json.Unmarshal([]byte("\""+vaccine+"\""), vt)
	if err != nil {
		return "", err
	}

	vs := &VaccinationSlot{
		VaccinationSlotData: VaccinationSlotData{
			Type: *vt,
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
	mySlot, err := readVaccinationSlot(ctx, mySlotUuid)
	if err != nil {
		return "", fmt.Errorf("slot: %s doesn't exist", mySlotUuid)
	}
	recipientSlot, err := readVaccinationSlot(ctx, recipientSlotUuid)
	if err != nil {
		return "", fmt.Errorf("slot: %s doesn't exist", recipientSlotUuid)
	}
	sender, err := getSender(ctx)
	if err != nil {
		return "", err
	}
	if sender != mySlot.Owner {
		return "", fmt.Errorf("%s doesn't own %s", sender, mySlotUuid)
	}
	if recipient != recipientSlot.Owner {
		return "", fmt.Errorf("%s doesn't own %s", recipient, recipientSlotUuid)
	}

	uuidWithHyphen := uuid.New()
	offerUuid = strings.Replace(uuidWithHyphen.String(), "-", "", -1)

	offer := TradeOffer{
		Uuid:          offerUuid,
		Sender:        sender,
		SenderItem:    mySlotUuid,
		Recipient:     recipient,
		RecipientItem: recipientSlotUuid,
	}

	err = offer.put(ctx)
	if err != nil {
		return "", err
	}

	return
}

func (c *VaccinationContract) AcceptOffer(ctx contractapi.TransactionContextInterface, offerUuid string) error {
	offer, err := getOffer(ctx, offerUuid)
	if err != nil {
		return err
	}
	recipient, err := getSender(ctx)
	if recipient != offer.Recipient {
		return fmt.Errorf("%s is not the recipient of the offer: %s", recipient, offerUuid)
	}
	senderSlot, err := readVaccinationSlot(ctx, offer.SenderItem)
	if err != nil {
		return err
	}
	recipientSlot, err := readVaccinationSlot(ctx, offer.RecipientItem)
	if err != nil {
		return err
	}
	if senderSlot.Owner != offer.Sender {
		return fmt.Errorf("sender: %s doesn't own the slot", senderSlot.Owner)
	}
	if recipientSlot.Owner != recipient {
		return fmt.Errorf("recipient: %s doesn't own the slot", recipient)
	}

	err = senderSlot.delBalance(ctx)
	if err != nil {
		return err
	}
	err = recipientSlot.delBalance(ctx)
	if err != nil {
		return err
	}

	senderSlot.Approved = ""
	recipientSlot.Approved = ""
	senderSlot.Owner, recipientSlot.Owner = recipientSlot.Owner, senderSlot.Owner
	senderSlot.Type, recipientSlot.Type = recipientSlot.Type, senderSlot.Type
	senderSlot.Previous, recipientSlot.Previous = recipientSlot.Previous, senderSlot.Previous

	err = senderSlot.put(ctx)
	if err != nil {
		return err
	}
	err = recipientSlot.put(ctx)
	if err != nil {
		return err
	}

	err = senderSlot.putBalance(ctx)
	if err != nil {
		return err
	}
	err = recipientSlot.putBalance(ctx)
	if err != nil {
		return err
	}

	err = c.emitTransfer(ctx, offer.Sender, offer.Recipient, offer.SenderItem)
	if err != nil {
		return err
	}
	err = c.emitTransfer(ctx, offer.Recipient, offer.Sender, offer.RecipientItem)
	if err != nil {
		return err
	}

	return nil
}

func (c *VaccinationContract) ListOffer(ctx contractapi.TransactionContextInterface) (string, error) {
	return "", nil
}

package chaincode

import (
	"encoding/base64"
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

func putOffer(ctx contractapi.TransactionContextInterface, offer TradeOffer) error {
	offerBytes, err := json.Marshal(&offer)
	if err != nil {
		return err
	}

	keySender, err := ctx.GetStub().CreateCompositeKey(offerPrefix, []string{offer.Sender, offer.Uuid})
	if err != nil {
		return err
	}
	keyRecipient, err := ctx.GetStub().CreateCompositeKey(offerPrefix, []string{offer.Recipient, offer.Uuid})
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(keySender, offerBytes)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(keyRecipient, offerBytes)
	if err != nil {
		return err
	}

	return nil
}

func getOffers(ctx contractapi.TransactionContextInterface, identity string) (offers []TradeOffer, err error) {
	offers = make([]TradeOffer, 0)

	iterator, err := ctx.GetStub().GetStateByPartialCompositeKey(offerPrefix, []string{identity})
	if err != nil {
		return
	}

	for iterator.HasNext() {
		offerKV, err := iterator.Next()
		if err != nil {
			return
		}
		offer := &TradeOffer{}
		err = json.Unmarshal(offerKV.Value, offer)
		if err != nil {
			return
		}
		offers = append(offers, *offer)
	}

	return
}

func delOffer(ctx contractapi.TransactionContextInterface, offer TradeOffer) error {
	keySender, err := ctx.GetStub().CreateCompositeKey(offerPrefix, []string{offer.Sender, offer.Uuid})
	if err != nil {
		return err
	}
	keyRecipient, err := ctx.GetStub().CreateCompositeKey(offerPrefix, []string{offer.Recipient, offer.Uuid})
	if err != nil {
		return err
	}

	err = ctx.GetStub().DelState(keySender)
	if err != nil {
		return err
	}

	err = ctx.GetStub().DelState(keyRecipient)
	if err != nil {
		return err
	}

	return nil
}

func getSender(ctx contractapi.TransactionContextInterface) (string, error) {
	sender64, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", fmt.Errorf("failed to get ClientIdentity: %v", err)
	}

	senderBytes, err := base64.StdEncoding.DecodeString(sender64)
	if err != nil {
		return "", fmt.Errorf("failed to decode sender64: %v", err)
	}
	sender := string(senderBytes)
	return sender, nil
}

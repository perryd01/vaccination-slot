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

	keyOffer, err := ctx.GetStub().CreateCompositeKey(offerPrefix, []string{offer.Uuid})
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

	err = ctx.GetStub().PutState(keyOffer, offerBytes)
	if err != nil {
		return err
	}

	return nil
}

func (offer TradeOffer) put(ctx contractapi.TransactionContextInterface) error {
	return putOffer(ctx, offer)
}

func getOffers(ctx contractapi.TransactionContextInterface, identity string) ([]TradeOffer, error) {
	offers := make([]TradeOffer, 0)

	iterator, err := ctx.GetStub().GetStateByPartialCompositeKey(offerPrefix, []string{identity})
	if err != nil {
		return offers, err
	}

	for iterator.HasNext() {
		offerKV, err := iterator.Next()
		if err != nil {
			return offers, err
		}
		offer := &TradeOffer{}
		err = json.Unmarshal(offerKV.Value, offer)
		if err != nil {
			return offers, err
		}
		offers = append(offers, *offer)
	}

	return offers, err
}

func getOffer(ctx contractapi.TransactionContextInterface, offerUuid string) (TradeOffer, error) {
	key, err := ctx.GetStub().CreateCompositeKey(offerPrefix, []string{offerUuid})
	offer := TradeOffer{}
	if err != nil {
		return offer, err
	}
	offerBytes, err := ctx.GetStub().GetState(key)
	if err != nil {
		return offer, err
	}
	err = json.Unmarshal(offerBytes, &offer)
	if err != nil {
		return offer, err
	}
	return offer, nil
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

	keyOffer, err := ctx.GetStub().CreateCompositeKey(offerPrefix, []string{offer.Uuid})
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

	err = ctx.GetStub().DelState(keyOffer)
	if err != nil {
		return err
	}

	return nil
}

func (offer TradeOffer) del(ctx contractapi.TransactionContextInterface) error {
	return delOffer(ctx, offer)
}

func (slot *VaccinationSlot) put(ctx contractapi.TransactionContextInterface) error {
	key, err := ctx.GetStub().CreateCompositeKey(vsPrefix, []string{slot.TokenId})
	if err != nil {
		return fmt.Errorf("failed to create CompositeKey: %v", err)
	}

	vsBytes, err := json.Marshal(slot)
	if err != nil {
		return fmt.Errorf("failed to marshal approval: %v", err)
	}

	err = ctx.GetStub().PutState(key, vsBytes)
	if err != nil {
		return fmt.Errorf("failed to PutState vsBytes: %v", err)
	}
	return nil
}

func (slot *VaccinationSlot) putBalance(ctx contractapi.TransactionContextInterface) error {
	key, err := ctx.GetStub().CreateCompositeKey(balancePrefix, []string{slot.Owner, slot.TokenId})
	if err != nil {
		return fmt.Errorf("failed to CreateCompositeKey: %v", err)
	}
	err = ctx.GetStub().PutState(key, []byte(slot.TokenId))
	if err != nil {
		return fmt.Errorf("failed to PutState balanceKeyTo %s: %v", key, err)
	}
	return nil
}

func (slot *VaccinationSlot) delBalance(ctx contractapi.TransactionContextInterface) error {
	key, err := ctx.GetStub().CreateCompositeKey(balancePrefix, []string{slot.Owner, slot.TokenId})
	if err != nil {
		return fmt.Errorf("failed to CreateCompositeKey: %v", err)
	}
	err = ctx.GetStub().DelState(key)
	if err != nil {
		return fmt.Errorf("failed to DelState balanceKeyFrom %s, %v", key, err)
	}
	return nil
}

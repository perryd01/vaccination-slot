package chaincode

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const (
	vsPrefix       = "nft"
	balancePrefix  = "balance"
	approvalPrefix = "approval"
	offerPrefix    = "offer"
)

func (c *VaccinationContract) emitTransfer(ctx contractapi.TransactionContextInterface, from, to, tokenId string) error {
	transferEvent := &Transfer{
		From:    from,
		To:      to,
		TokenId: tokenId,
	}

	transferEventBytes, err := json.Marshal(transferEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal transferEvent: %v", err)
	}

	err = ctx.GetStub().SetEvent("Transfer", transferEventBytes)
	if err != nil {
		return fmt.Errorf("failed to SetEvent transformEventBytes %s: %v", transferEventBytes, err)
	}

	return nil
}

func (c *VaccinationContract) emitApproval(ctx contractapi.TransactionContextInterface, owner, approved, tokenId string) error {
	approval := &Approval{
		Owner:    owner,
		Approved: approved,
		TokenId:  tokenId,
	}

	approvalBytes, err := json.Marshal(approval)
	if err != nil {
		return err
	}

	err = ctx.GetStub().SetEvent("Approval", approvalBytes)
	if err != nil {
		return err
	}
	return nil
}

func (c *VaccinationContract) emitApprovalForAll(ctx contractapi.TransactionContextInterface, owner, operator string, approved bool) error {
	approval := &ApprovalForAll{
		Owner:    owner,
		Operator: operator,
		Approved: approved,
	}

	approvalBytes, err := json.Marshal(approval)
	if err != nil {
		return err
	}

	err = ctx.GetStub().SetEvent("ApprovalForAll", approvalBytes)
	if err != nil {
		return err
	}
	return nil
}

func (c *VaccinationContract) BalanceOf(ctx contractapi.TransactionContextInterface, owner string) int {
	owner64 := base64.StdEncoding.EncodeToString([]byte(owner))
	iterator, err := ctx.GetStub().GetStateByPartialCompositeKey(balancePrefix, []string{owner64})
	if err != nil {
		panic("Error creating asset chaincode: " + err.Error())
	}

	balance := 0
	for iterator.HasNext() {
		_, err := iterator.Next()
		if err != nil {
			return 0
		}
		balance++
	}
	return balance
}

func (c *VaccinationContract) OwnerOf(ctx contractapi.TransactionContextInterface, tokenId string) (string, error) {
	vs, err := readVaccinationSlot(ctx, tokenId)
	if err != nil {
		return "", err
	}
	return vs.Owner, nil
}

func (c *VaccinationContract) TransferFrom(ctx contractapi.TransactionContextInterface, from string, to string, tokenId string) (bool, error) {
	sender64, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return false, fmt.Errorf("failed to get ClientIdentity: %v", err)
	}

	senderBytes, err := base64.StdEncoding.DecodeString(sender64)
	if err != nil {
		return false, fmt.Errorf("failed to decode sender64: %v", err)
	}
	sender := string(senderBytes)

	vs, err := readVaccinationSlot(ctx, tokenId)
	if err != nil {
		return false, fmt.Errorf("failed to read vaccination slot: %v", err)
	}

	owner := vs.Owner
	operator := vs.Approved
	operatorApproval, err := c.IsApprovedForAll(ctx, owner, sender)
	if err != nil {
		return false, fmt.Errorf("failed to get IsApprovedForAll: %v", err)
	}
	if owner != sender && operator != sender && !operatorApproval {
		return false, fmt.Errorf("the sender is not the current owner nor an authorized operator")
	}

	if owner != from {
		return false, fmt.Errorf("the from is not the current owner")
	}

	vs.Approved = ""

	vs.Owner = to
	key, err := ctx.GetStub().CreateCompositeKey(vsPrefix, []string{tokenId})
	if err != nil {
		return false, fmt.Errorf("failed to create CompositeKey: %v", err)
	}

	vsBytes, err := json.Marshal(vs)
	if err != nil {
		return false, fmt.Errorf("failed to marshal approval: %v", err)
	}

	err = ctx.GetStub().PutState(key, vsBytes)
	if err != nil {
		return false, fmt.Errorf("failed to PutState vsBytes: %v", err)
	}

	from64 := base64.StdEncoding.EncodeToString([]byte(from))
	balanceKeyFrom, err := ctx.GetStub().CreateCompositeKey(balancePrefix, []string{from64, tokenId})
	if err != nil {
		return false, fmt.Errorf("failed to CreateCompositeKey: %v", err)
	}

	err = ctx.GetStub().DelState(balanceKeyFrom)
	if err != nil {
		return false, fmt.Errorf("failed to DelState balanceKeyFrom %s, %v", balanceKeyFrom, err)
	}

	to64 := base64.StdEncoding.EncodeToString([]byte(to))
	balanceKeyTo, err := ctx.GetStub().CreateCompositeKey(balancePrefix, []string{to64, tokenId})
	if err != nil {
		return false, fmt.Errorf("failed to create CompositeKey: %v", err)
	}

	err = ctx.GetStub().PutState(balanceKeyTo, []byte(tokenId))
	if err != nil {
		return false, fmt.Errorf("failed to PutState balanceKeyTo %s: %v", balanceKeyTo, err)
	}

	err = c.emitTransfer(ctx, from, to, tokenId)
	if err != nil {
		return false, nil
	}

	return true, nil
}

func (c *VaccinationContract) Approve(ctx contractapi.TransactionContextInterface, operator string, tokenId string) (bool, error) {
	sender64, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return false, fmt.Errorf("failed to get client identity: %v", err)
	}

	senderBytes, err := base64.StdEncoding.DecodeString(sender64)
	if err != nil {
		return false, fmt.Errorf("failed to decode sender64: %v", err)
	}
	sender := string(senderBytes)

	vs, err := readVaccinationSlot(ctx, tokenId)
	if err != nil {
		return false, fmt.Errorf("failed to read vaccination slot: %v", err)
	}

	owner := vs.Owner
	operatorApproval, err := c.IsApprovedForAll(ctx, owner, sender)
	if err != nil {
		return false, fmt.Errorf("failed to get IsApprovedForAll: %v", err)
	}
	if owner != sender && !operatorApproval {
		return false, fmt.Errorf("the sender is not the current owner nor an authorized oeprator")
	}

	vs.Approved = operator
	key, err := ctx.GetStub().CreateCompositeKey(vsPrefix, []string{tokenId})
	if err != nil {
		return false, fmt.Errorf("failed to create CompositeKey %s: %v", tokenId, err)
	}

	vsBytes, err := json.Marshal(vs)
	if err != nil {
		return false, fmt.Errorf("failed to marshal vsBytes: %v", err)
	}

	err = ctx.GetStub().PutState(key, vsBytes)
	if err != nil {
		return false, fmt.Errorf("failed to PutState for key: %v", err)
	}

	err = c.emitApproval(ctx, vs.Owner, operator, tokenId)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (c *VaccinationContract) SetApprovalForAll(ctx contractapi.TransactionContextInterface, operator string, approved bool) (bool, error) {
	sender64, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return false, fmt.Errorf("failed to get client identity: %v", err)
	}

	senderBytes, err := base64.StdEncoding.DecodeString(sender64)
	if err != nil {
		return false, fmt.Errorf("failed to decode sender64: %v", err)
	}
	sender := string(senderBytes)

	vsApproval := &ApprovalForAll{
		Owner:    sender,
		Operator: operator,
		Approved: approved,
	}

	approvalKey, err := ctx.GetStub().CreateCompositeKey(approvalPrefix, []string{sender, operator})
	if err != nil {
		return false, fmt.Errorf("failed to create CompositeKey: %v", err)
	}

	approvalBytes, err := json.Marshal(vsApproval)
	if err != nil {
		return false, fmt.Errorf("failed to marshal vsApproval: %v", err)
	}

	err = ctx.GetStub().PutState(approvalKey, approvalBytes)
	if err != nil {
		return false, fmt.Errorf("failed to putState approvalBytes: %v", err)
	}

	err = c.emitApprovalForAll(ctx, sender, operator, approved)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (c *VaccinationContract) GetApproved(ctx contractapi.TransactionContextInterface, tokenId string) (string, error) {
	vs, err := readVaccinationSlot(ctx, tokenId)
	if err != nil {
		return "", fmt.Errorf("failed GetApproved for tokenId: %v", err)
	}
	return vs.Approved, nil
}

func (c *VaccinationContract) IsApprovedForAll(ctx contractapi.TransactionContextInterface, owner string, operator string) (bool, error) {
	approvalKey, err := ctx.GetStub().CreateCompositeKey(approvalPrefix, []string{owner, operator})
	if err != nil {
		return false, fmt.Errorf("failed to create CompositeKey: %v", err)
	}
	approvalBytes, err := ctx.GetStub().GetState(approvalKey)
	if err != nil {
		return false, fmt.Errorf("failed to GetState approvalBytes: %v", err)
	}

	if len(approvalBytes) < 1 {
		return false, nil
	}

	approval := &ApprovalForAll{}
	err = json.Unmarshal(approvalBytes, approval)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal: %v, string %s", err, string(approvalBytes))
	}

	return approval.Approved, nil
}

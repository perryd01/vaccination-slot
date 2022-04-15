package chaincode

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// Define objectType names for prefix
const balancePrefix = "balance"
const tokenPrefix = "token"
const approvalPrefix = "approval"

// Define key names for options
const nameKey = "name"
const symbolKey = "symbol"

const hospitalMSPID = "hosp1"

type VSTokenContract struct {
	contractapi.Contract
}

func _readToken(ctx contractapi.TransactionContextInterface, tokenId string) (*Token, error) {
	tokenKey, err := ctx.GetStub().CreateCompositeKey(tokenPrefix, []string{tokenId})
	if err != nil {
		return nil, fmt.Errorf("failed to CreateCompositeKey %s: %v", tokenId, err)
	}
	tokenBytes, err := ctx.GetStub().GetState(tokenKey)
	if err != nil {
		return nil, fmt.Errorf("failed to GetState %s: %v", tokenId, err)
	}

	token := &Token{}
	err = json.Unmarshal(tokenBytes, token)
	if err != nil {
		return nil, fmt.Errorf("failed to Unmarshal %v", err)
	}

	return token, nil
}

func _tokenExists(ctx contractapi.TransactionContextInterface, tokenId string) bool {
	tokenKey, err := ctx.GetStub().CreateCompositeKey(tokenPrefix, []string{tokenId})
	if err != nil {
		panic("error creating CreateCompositeKey:" + err.Error())
	}

	tokenBytes, err := ctx.GetStub().GetState(tokenKey)
	if err != nil {
		panic("error GetState:" + err.Error())
	}

	return len(tokenBytes) > 0
}

// BalanceOf counts all non-fungible tokens assigned to an owner
// param owner {String} An owner for whom to query the balance
// returns {int} The number of non-fungible tokens owned by the owner, possibly zero
func (c *VSTokenContract) BalanceOf(ctx contractapi.TransactionContextInterface, owner string) int {
	// There is a key record for every non-fungible token in the format of balancePrefix.owner.tokenId.
	// BalanceOf() queries for and counts all records matching balancePrefix.owner.*

	iterator, err := ctx.GetStub().GetStateByPartialCompositeKey(balancePrefix, []string{owner})
	if err != nil {
		panic("Error creating asset chaincode:" + err.Error())
	}

	// Count the number of returned composite keys
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

// OwnerOf finds the owner of a non-fungible token
// param {String} tokenId The identifier for a non-fungible token
// returns {String} Return the owner of the non-fungible token
func (c *VSTokenContract) OwnerOf(ctx contractapi.TransactionContextInterface, tokenId string) (string, error) {
	nft, err := _readToken(ctx, tokenId)
	if err != nil {
		return "", fmt.Errorf("could not process OwnerOf for tokenId: %w", err)
	}

	return nft.Owner, nil
}

// Approve changes or reaffirms the approved client for a non-fungible token
// param {String} operator The new approved client
// param {String} tokenId the non-fungible token to approve
// returns {Boolean} Return whether the approval was successful or not
func (c *VSTokenContract) Approve(ctx contractapi.TransactionContextInterface, operator string, tokenId string) (bool, error) {
	sender64, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return false, fmt.Errorf("failed to GetClientIdentity: %v", err)
	}

	senderBytes, err := base64.StdEncoding.DecodeString(sender64)
	if err != nil {
		return false, fmt.Errorf("failed to DecodeString senderBytes: %v", err)
	}
	sender := string(senderBytes)

	token, err := _readToken(ctx, tokenId)
	if err != nil {
		return false, fmt.Errorf("failed to _readNFT: %v", err)
	}

	// Check if the sender is the current owner of the non-fungible token
	// or an authorized operator of the current owner
	owner := token.Owner
	operatorApproval, err := c.IsApprovedForAll(ctx, owner, sender)
	if err != nil {
		return false, fmt.Errorf("failed to get IsApprovedForAll: %v", err)
	}
	if owner != sender && !operatorApproval {
		return false, fmt.Errorf("the sender is not the current owner nor an authorized operator")
	}

	// Update the approved operator of the non-fungible token
	token.Approved = operator
	tokenKey, err := ctx.GetStub().CreateCompositeKey(tokenPrefix, []string{tokenId})
	if err != nil {
		return false, fmt.Errorf("failed to CreateCompositeKey %s: %v", tokenKey, err)
	}

	tokenBytes, err := json.Marshal(token)
	if err != nil {
		return false, fmt.Errorf("failed to marshal tokenBytes: %v", err)
	}

	err = ctx.GetStub().PutState(tokenKey, tokenBytes)
	if err != nil {
		return false, fmt.Errorf("failed to PutState for tokenKey: %v", err)
	}

	return true, nil
}

// SetApprovalForAll enables or disables approval for a third party ("operator")
// to manage all the message sender's assets
// param {String} operator A client to add to the set of authorized operators
// param {Boolean} approved True if the operator is approved, false to revoke approval
// returns {Boolean} Return whether the approval was successful or not
func (c *VSTokenContract) SetApprovalForAll(ctx contractapi.TransactionContextInterface, operator string, approved bool) (bool, error) {
	sender64, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return false, fmt.Errorf("failed to GetClientIdentity: %v", err)
	}

	senderBytes, err := base64.StdEncoding.DecodeString(sender64)
	if err != nil {
		return false, fmt.Errorf("failed to DecodeString sender: %v", err)
	}
	sender := string(senderBytes)

	tokenApproval := new(Approval)
	tokenApproval.Owner = sender
	tokenApproval.Operator = operator
	tokenApproval.Approved = approved

	approvalKey, err := ctx.GetStub().CreateCompositeKey(approvalPrefix, []string{sender, operator})
	if err != nil {
		return false, fmt.Errorf("failed to CreateCompositeKey: %v", err)
	}

	approvalBytes, err := json.Marshal(tokenApproval)
	if err != nil {
		return false, fmt.Errorf("failed to marshal approvalBytes: %v", err)
	}

	err = ctx.GetStub().PutState(approvalKey, approvalBytes)
	if err != nil {
		return false, fmt.Errorf("failed to PutState approvalBytes: %v", err)
	}

	// Emit the ApprovalForAll event
	err = ctx.GetStub().SetEvent("ApprovalForAll", approvalBytes)
	if err != nil {
		return false, fmt.Errorf("failed to SetEvent ApprovalForAll: %v", err)
	}

	return true, nil
}

// IsApprovedForAll returns if a client is an authorized operator for another client
// param {String} owner The client that owns the non-fungible tokens
// param {String} operator The client that acts on behalf of the owner
// returns {Boolean} Return true if the operator is an approved operator for the owner, false otherwise
func (c *VSTokenContract) IsApprovedForAll(ctx contractapi.TransactionContextInterface, owner string, operator string) (bool, error) {
	approvalKey, err := ctx.GetStub().CreateCompositeKey(approvalPrefix, []string{owner, operator})
	if err != nil {
		return false, fmt.Errorf("failed to CreateCompositeKey: %v", err)
	}
	approvalBytes, err := ctx.GetStub().GetState(approvalKey)
	if err != nil {
		return false, fmt.Errorf("failed to GetState approvalBytes %s: %v", approvalBytes, err)
	}

	if len(approvalBytes) < 1 {
		return false, nil
	}

	approval := new(Approval)
	err = json.Unmarshal(approvalBytes, approval)
	if err != nil {
		return false, fmt.Errorf("failed to Unmarshal: %v, string %s", err, string(approvalBytes))
	}

	return approval.Approved, nil

}

// GetApproved returns the approved client for a single non-fungible token
// param {String} tokenId the non-fungible token to find the approved client for
// returns {Object} Return the approved client for this non-fungible token, or null if there is none
func (c *VSTokenContract) GetApproved(ctx contractapi.TransactionContextInterface, tokenId string) (string, error) {
	nft, err := _readToken(ctx, tokenId)
	if err != nil {
		return "false", fmt.Errorf("failed GetApproved for tokenId : %v", err)
	}
	return nft.Approved, nil
}

// TransferFrom transfers the ownership of a non-fungible token
// from one owner to another owner
// param {String} from The current owner of the non-fungible token
// param {String} to The new owner
// param {String} tokenId the non-fungible token to transfer
// returns {Boolean} Return whether the transfer was successful or not

func (c *VSTokenContract) TransferFrom(ctx contractapi.TransactionContextInterface, from string, to string, tokenId string) (bool, error) {

	// Get ID of submitting client identity
	sender64, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return false, fmt.Errorf("failed to GetClientIdentity: %v", err)
	}

	senderBytes, err := base64.StdEncoding.DecodeString(sender64)
	if err != nil {
		return false, fmt.Errorf("failed to DecodeString sender: %v", err)
	}
	sender := string(senderBytes)

	token, err := _readToken(ctx, tokenId)
	if err != nil {
		return false, fmt.Errorf("failed to _readNFT : %v", err)
	}

	owner := token.Owner
	operator := token.Approved
	operatorApproval, err := c.IsApprovedForAll(ctx, owner, sender)
	if err != nil {
		return false, fmt.Errorf("failed to get IsApprovedForAll : %v", err)
	}
	if owner != sender && operator != sender && !operatorApproval {
		return false, fmt.Errorf("the sender is not the current owner nor an authorized operator")
	}

	// Check if `from` is the current owner
	if owner != from {
		return false, fmt.Errorf("the from is not the current owner")
	}

	// Clear the approved client for this non-fungible token
	token.Approved = ""

	// Overwrite a non-fungible token to assign a new owner.
	token.Owner = to
	tokenKey, err := ctx.GetStub().CreateCompositeKey(tokenPrefix, []string{tokenId})
	if err != nil {
		return false, fmt.Errorf("failed to CreateCompositeKey: %v", err)
	}

	tokenBytes, err := json.Marshal(token)
	if err != nil {
		return false, fmt.Errorf("failed to marshal approval: %v", err)
	}

	err = ctx.GetStub().PutState(tokenKey, tokenBytes)
	if err != nil {
		return false, fmt.Errorf("failed to PutState nftBytes %s: %v", tokenBytes, err)
	}

	// Remove a composite key from the balance of the current owner
	balanceKeyFrom, err := ctx.GetStub().CreateCompositeKey(balancePrefix, []string{from, tokenId})
	if err != nil {
		return false, fmt.Errorf("failed to CreateCompositeKey from: %v", err)
	}

	err = ctx.GetStub().DelState(balanceKeyFrom)
	if err != nil {
		return false, fmt.Errorf("failed to DelState balanceKeyFrom %s: %v", tokenBytes, err)
	}

	// Save a composite key to count the balance of a new owner
	balanceKeyTo, err := ctx.GetStub().CreateCompositeKey(balancePrefix, []string{to, tokenId})
	if err != nil {
		return false, fmt.Errorf("failed to CreateCompositeKey to: %v", err)
	}
	err = ctx.GetStub().PutState(balanceKeyTo, []byte{0})
	if err != nil {
		return false, fmt.Errorf("failed to PutState balanceKeyTo %s: %v", balanceKeyTo, err)
	}

	// Emit the Transfer event
	transferEvent := new(Transfer)
	transferEvent.From = from
	transferEvent.To = to
	transferEvent.TokenId = tokenId

	transferEventBytes, err := json.Marshal(transferEvent)
	if err != nil {
		return false, fmt.Errorf("failed to marshal transferEventBytes: %v", err)
	}

	err = ctx.GetStub().SetEvent("Transfer", transferEventBytes)
	if err != nil {
		return false, fmt.Errorf("failed to SetEvent transferEventBytes %s: %v", transferEventBytes, err)
	}
	return true, nil
}

func (c *VSTokenContract) Name(ctx contractapi.TransactionContextInterface) (string, error) {
	bytes, err := ctx.GetStub().GetState(nameKey)
	if err != nil {
		return "", fmt.Errorf("failed to get Name bytes: %s", err)
	}

	return string(bytes), nil
}

func (c *VSTokenContract) Symbol(ctx contractapi.TransactionContextInterface) (string, error) {
	bytes, err := ctx.GetStub().GetState(symbolKey)
	if err != nil {
		return "", fmt.Errorf("failed to get Symbol: %v", err)
	}

	return string(bytes), nil
}

func (c *VSTokenContract) CreateToken(ctx contractapi.TransactionContextInterface, tokenId string, tokenType string, date time.Time) (*Token, error) {
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return nil, fmt.Errorf("failed to get clientMSPID: %v", err)
	}

	if clientMSPID != hospitalMSPID {
		return nil, fmt.Errorf("client is not authorized to set the name and symbol of the token")
	}

	creator64, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return nil, fmt.Errorf("failed to get minter id: %v", err)
	}

	creatorBytes, err := base64.StdEncoding.DecodeString(creator64)
	if err != nil {
		return nil, fmt.Errorf("failed to DecodeString minter64: %v", err)
	}
	creator := string(creatorBytes)

	exists := _tokenExists(ctx, tokenId)
	if exists {
		return nil, fmt.Errorf("the token %s is already minted.: %v", tokenId, err)
	}

	vstoken := &Token{
		TokenId: tokenId,
		Owner:   creator,
		Type:    tokenType,
		Burned:  false,
		Date:    date,
	}

	tokenKey, err := ctx.GetStub().CreateCompositeKey(tokenPrefix, []string{tokenId})
	if err != nil {
		return nil, fmt.Errorf("failed to CreateCompositeKey to tokenKey: %v", err)
	}

	tokenBytes, err := json.Marshal(vstoken)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal token: %v", err)
	}

	err = ctx.GetStub().PutState(tokenKey, tokenBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to PutState nftBytes %s: %v", tokenBytes, err)
	}

	balanceKey, err := ctx.GetStub().CreateCompositeKey(balancePrefix, []string{creator, tokenId})
	if err != nil {
		return nil, fmt.Errorf("failed to CreateCompositeKey to balanceKey: %v", err)
	}

	err = ctx.GetStub().PutState(balanceKey, []byte{'\u0000'})
	if err != nil {
		return nil, fmt.Errorf("failed to PutState balanceKey %s: %v", tokenBytes, err)
	}

	// Emit the Transfer event
	transferEvent := new(Transfer)
	transferEvent.From = "0x0"
	transferEvent.To = creator
	transferEvent.TokenId = tokenId

	transferEventBytes, err := json.Marshal(transferEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal transferEventBytes: %v", err)
	}

	err = ctx.GetStub().SetEvent("Transfer", transferEventBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to SetEvent transferEventBytes %s: %v", transferEventBytes, err)
	}

	return vstoken, nil
}

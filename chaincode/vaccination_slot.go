// Package chaincode implements ERC721 for HF9 Vaccination Slots.
package chaincode

// VaccinationSlotData contains information about specific occasion
type VaccinationSlotData struct {
	// Type of the vaccine.
	// May change when the token is transferred.
	Type VaccinationType `json:"type"`

	// When the vaccine should be administered.
	// Never changes.
	Date VaccinationDate `json:"date"`

	// Previously administered vaccine of the same type.
	// If present it may forbid the transfer of the token.
	// May change when the token is transferred.
	Previous string `json:"previous,omitempty"`
}

// VaccinationSlot contains ERC712 related data (this is the NFT)
type VaccinationSlot struct {
	VaccinationSlotData
	TokenId  string `json:"tokenId"`
	Owner    string `json:"owner"`
	Approved string `json:"approved"`
}

type Approval struct {
	Owner    string `json:"owner"`
	Approved string `json:"operator"`
	TokenId  string `json:"tokenId"`
}

type ApprovalForAll struct {
	Owner    string `json:"owner"`
	Operator string `json:"operator"`
	Approved bool   `json:"approved"`
}

type Transfer struct {
	From    string `json:"from"`
	To      string `json:"to"`
	TokenId string `json:"tokenId"`
}

// TradeOffer represents a trade offer for specific slots of specific identities.
//
// Making an offer:
//  func (c *VaccinationContract) MakeOffer(ctx contractapi.TransactionContextInterface, mySlotUuid, recipient, recipientSlotUuid string) (offerUuid string, err error)
// Accepting an offer:
//  func (c *VaccinationContract) AcceptOffer(ctx contractapi.TransactionContextInterface, offerUuid string) error
// Offers are stored in the global state as
// offer.sender.offerUuid and offer.recipient.offerUuid offer.offerUuid.
// This way it enables queries by partial key.
type TradeOffer struct {
	Uuid          string `json:"uuid"`
	Sender        string `json:"sender"`
	SenderItem    string `json:"senderItem"`
	Recipient     string `json:"recipient"`
	RecipientItem string `json:"recipientItem"`
}

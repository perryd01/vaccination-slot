// Package chaincode implements ERC721 for HF9 Vaccination Slots.
package chaincode

// VaccinationSlotData contains information about specific occasion
type VaccinationSlotData struct {
	// Type of the vaccine.
	// May change when the token is transferred.
	Type string `json:"type"`

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
	Operator string `json:"operator"`
	Approved bool   `json:"approved"`
}

type Transfer struct {
	From    string `json:"from"`
	To      string `json:"to"`
	TokenId string `json:"tokenId"`
}

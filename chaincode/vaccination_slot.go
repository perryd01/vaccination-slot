package chaincode

type Token struct {
	TokenId  string          `json:"tokenId"`
	Type     string          `json:"type"`
	Date     VaccinationDate `json:"date"`
	Owner    string          `json:"owner"`
	Burned   bool            `json:"approved"`
	Approved string          `json:"approved"`
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

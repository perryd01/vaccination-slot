package chaincode

import "github.com/hyperledger/fabric-contract-api-go/contractapi"

func (c *VSTokenContract) Type(ctx contractapi.TransactionContextInterface, tokenId string) (string, error) {
	token, err := _readToken(ctx, tokenId)
	if err != nil {
		return "", err
	}
	s, err := token.Type.MarshalJSON()
	if err != nil {
		return "", err
	}
	return string(s), nil
}

func (c *VSTokenContract) IsBurned(ctx contractapi.TransactionContextInterface, tokenId string) (bool, error) {
	token, err := _readToken(ctx, tokenId)
	if err != nil {
		return false, err
	}
	return token.Burned, nil
}

func (c *VSTokenContract) Date(ctx contractapi.TransactionContextInterface, tokenId string) (string, error) {
	token, err := _readToken(ctx, tokenId)
	if err != nil {
		return "", err
	}
	s, err := token.Date.MarshalJSON()
	// TODO review if works
	return string(s), nil
}

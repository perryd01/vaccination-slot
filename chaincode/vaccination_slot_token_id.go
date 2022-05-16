package chaincode

import (
	"github.com/google/uuid"
	"strings"
)

type TokenIdGeneratorInterface interface {
	HasNext() bool
	Next() string
}

type TokenIdGenerator struct {
}

func (g *TokenIdGenerator) HasNext() bool {
	return true
}

func (g *TokenIdGenerator) Next() string {
	uuidWithHyphen := uuid.New()
	tokenUuid := strings.Replace(uuidWithHyphen.String(), "-", "", -1)
	return tokenUuid
}

package chaincode

import (
	"testing"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
	"github.com/stretchr/testify/mock"
)

type MockStub struct {
	shim.ChaincodeStubInterface
	mock.Mock
}

type MockClientIdentity struct {
	cid.ClientIdentity
	mock.Mock
}

type MockContext struct {
	contractapi.TransactionContextInterface
	mock.Mock
}

type MockIterator struct {
	shim.StateQueryIteratorInterface
	queryresult.KV
}

func (ms *MockStub) GetStateByPartialCompositeKey(objectType string, keys []string) (shim.StateQueryIteratorInterface, error) {
	args := ms.Called(objectType, keys)
	return args.Get(0).(shim.StateQueryIteratorInterface), args.Error(1)
}

func (ms *MockStub) GetState(key string) ([]byte, error) {
	args := ms.Called(key)
	return args.Get(0).([]byte), args.Error(1)
}

func (ms *MockStub) PutState(key string, value []byte) error {
	args := ms.Called(key, value)
	return args.Error(0)
}
func (ms *MockStub) SetEvent(key string, value []byte) error {
	args := ms.Called(key, value)
	return args.Error(0)
}

func (ms *MockStub) DelState(key string) error {
	args := ms.Called(key)
	return args.Error(0)
}

func (ms *MockStub) CreateCompositeKey(objectType string, attributes []string) (string, error) {
	args := ms.Called(objectType, attributes)
	return args.Get(0).(string), args.Error(1)
}

func (mci *MockClientIdentity) GetID() (string, error) {
	args := mci.Called()
	return args.Get(0).(string), args.Error(1)
}

func (mci *MockClientIdentity) GetMSPID() (string, error) {
	args := mci.Called()
	return args.Get(0).(string), args.Error(1)
}

func (mc *MockContext) GetStub() shim.ChaincodeStubInterface {
	args := mc.Called()
	return args.Get(0).(*MockStub)
}

func (mc *MockContext) GetClientIdentity() cid.ClientIdentity {
	args := mc.Called()
	return args.Get(0).(*MockClientIdentity)
}

func setupStub() (*MockContext, *MockStub) {

	ms := new(MockStub)
	mc := new(MockContext)

	return mc, ms
}

func (it *MockIterator) HasNext() bool {
	return false
}

func TestTokenMinting(t *testing.T) {

}

func TestUnauthorizedTokenMinting(t *testing.T) {

}

func TestSuccessfulTrading(t *testing.T) {}

func TestTradingBurnedToken(t *testing.T) {}

func TestTradingTokenWithInvalidDate(t *testing.T) {}

func TestTradingNonExistentToken(t *testing.T) {}

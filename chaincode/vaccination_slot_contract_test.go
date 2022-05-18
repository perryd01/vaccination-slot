package chaincode

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"log"
	"strings"
	"testing"
	"time"
)

const (
	doctor1  = "x509::CN=Doctor1,OU=client::CN=MedicalStation CA"
	patient1 = "x509::CN=Patient1,OU=client::CN=Patients CA"
	patient2 = "x509::CN=Patient2,OU=client::CN=Patients CA"
	slot1    = "slot1"
	slot2    = "slot2"
	slot3    = "slot3"
	slot4    = "slot4"
	offer1   = "offer1"
	offer2   = "offer2"
)

const (
	getStub                       = "GetStub"
	createCompositeKey            = "CreateCompositeKey"
	getState                      = "GetState"
	putState                      = "PutState"
	getStateByPartialCompositeKey = "GetStateByPartialCompositeKey"
	getClientIdentity             = "GetClientIdentity"
	setEvent                      = "SetEvent"
	getMSPID                      = "GetMSPID"
	getID                         = "GetID"
	delState                      = "DelState"
)

type MockStub struct {
	shim.ChaincodeStubInterface
	mock.Mock
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

type MockClientIdentity struct {
	cid.ClientIdentity
	mock.Mock
}

func (mci *MockClientIdentity) GetID() (string, error) {
	args := mci.Called()
	return args.Get(0).(string), args.Error(1)
}

func (mci *MockClientIdentity) GetMSPID() (string, error) {
	args := mci.Called()
	return args.Get(0).(string), args.Error(1)
}

type MockContext struct {
	contractapi.TransactionContextInterface
	mock.Mock
}

func (mc *MockContext) GetStub() shim.ChaincodeStubInterface {
	args := mc.Called()
	return args.Get(0).(*MockStub)
}

func (mc *MockContext) GetClientIdentity() cid.ClientIdentity {
	args := mc.Called()
	return args.Get(0).(*MockClientIdentity)
}

type MockIterator struct {
	shim.StateQueryIteratorInterface
	//queryresult.KV
	queries []queryresult.KV
}

func (it *MockIterator) HasNext() bool {
	return len(it.queries) > 0
}

func (it *MockIterator) Next() (*queryresult.KV, error) {
	if it.HasNext() {
		value := it.queries[0]
		it.queries = it.queries[1:]
		return &value, nil
	}
	return nil, nil
}

type MockTokenIdGenerator struct {
	Ids []string
}

func (g *MockTokenIdGenerator) HasNext() bool {
	return len(g.Ids) > 0
}

func (g *MockTokenIdGenerator) Next() string {
	if g.HasNext() {
		id := g.Ids[0]
		g.Ids = g.Ids[1:]
		return id
	}
	return ""
}

//<editor-fold desc="Test BalanceOf">
func TestBalanceOf(t *testing.T) {
	ctx := setupTestBalanceOf()
	c := &VaccinationContract{}

	balance := c.BalanceOf(ctx, patient1)
	assert.Equal(t, 0, balance)

	balance = c.BalanceOf(ctx, patient2)
	assert.Equal(t, 2, balance)
}

func setupTestBalanceOf() *MockContext {
	ms := &MockStub{}
	emptyIterator := &MockIterator{}
	iterator := &MockIterator{
		queries: []queryresult.KV{
			{
				Key:   "Igen",
				Value: []byte("Igen"),
			},
			{
				Key:   "Nem",
				Value: []byte("Nem"),
			},
		},
	}

	patient164 := base64.StdEncoding.EncodeToString([]byte(patient1))
	patient264 := base64.StdEncoding.EncodeToString([]byte(patient2))

	ms.On(getStateByPartialCompositeKey, balancePrefix, []string{patient164}).Return(emptyIterator, nil)
	ms.On(getStateByPartialCompositeKey, balancePrefix, []string{patient264}).Return(iterator, nil)
	ms.On(getStateByPartialCompositeKey, vsPrefix, []string{}).Return(emptyIterator, nil)

	mci := &MockClientIdentity{}

	mc := &MockContext{}
	mc.On(getStub).Return(ms)
	mc.On(getClientIdentity).Return(mci)
	return mc
}

//</editor-fold>

//<editor-fold desc="Test OwnerOf">
func TestOwnerOf(t *testing.T) {
	ctx := setupTestOwnerOf()
	c := &VaccinationContract{}

	owner, _ := c.OwnerOf(ctx, "slot1")
	assert.Equal(t, patient1, owner)
}

func setupTestOwnerOf() *MockContext {
	ms := &MockStub{}

	vs := &VaccinationSlot{
		VaccinationSlotData: VaccinationSlotData{
			Type: Delta,
			Date: VaccinationDate(time.Now()),
		},
		TokenId: "slot1",
		Owner:   patient1,
	}
	vsb, _ := json.Marshal(vs)

	ms.On(createCompositeKey, vsPrefix, []string{"slot1"}).Return(vsPrefix+".slot1", nil)
	ms.On(getState, vsPrefix+".slot1").Return(vsb, nil)

	mci := &MockClientIdentity{}

	mc := &MockContext{}
	mc.On(getStub).Return(ms)
	mc.On(getClientIdentity).Return(mci)

	return mc
}

//</editor-fold>

//<editor-fold desc="Test IssueSlot">
func TestIssueSlot(t *testing.T) {
	t.Run("Correct", func(t *testing.T) {
		ctx, ms, gen := setupTestIssueSlot1()
		c := &VaccinationContract{
			IdGenerator: gen,
		}
		slot1, err := c.IssueSlot(ctx, "delta", "2000-01-01", patient1, "")
		assert.Equal(t, nil, err)
		assert.NotEmpty(t, slot1)
		ms.AssertCalled(t, setEvent, "Transfer", mock.AnythingOfType("[]uint8"))
	})
	t.Run("Wrong MSPID", func(t *testing.T) {
		ctx, ms, gen := setupTestIssueSlot2()
		c := &VaccinationContract{
			IdGenerator: gen,
		}
		slot1, err := c.IssueSlot(ctx, "delta", "2000-01-01", patient1, "")
		assert.Error(t, err)
		assert.Empty(t, slot1)
		ms.AssertNotCalled(t, setEvent, "Transfer", mock.Anything)
	})
	t.Run("Occupied", func(t *testing.T) {
		ctx, ms, gen := setupTestIssueSlot3()
		c := &VaccinationContract{
			IdGenerator: gen,
		}
		_, err := c.IssueSlot(ctx, "delta", "2000-01-01", patient1, "")
		assert.Error(t, err)
		ms.AssertNotCalled(t, setEvent, "Transfer", mock.Anything)
	})
	//t.Run("Wrong vaccine", func(t *testing.T) {
	//	ctx, ms, gen := setupTestIssueSlot1()
	//	c := &VaccinationContract{
	//		IdGenerator: gen,
	//	}
	//	_, err := c.IssueSlot(ctx, "macskakaja", "2000-01-01", patient1)
	//	assert.Error(t, err)
	//	ms.AssertNotCalled(t, "SetEvent", mock.Anything)
	//})
}

func setupTestIssueSlot1() (*MockContext, *MockStub, *MockTokenIdGenerator) {
	ms := &MockStub{}
	gen := &MockTokenIdGenerator{
		[]string{slot1, slot2},
	}

	anyBytes := mock.AnythingOfType("[]uint8")

	patient1Balance := &MockIterator{}
	patient164 := base64.StdEncoding.EncodeToString([]byte(patient1))
	ms.On(getStateByPartialCompositeKey, balancePrefix, []string{patient164}).Return(patient1Balance, nil)
	{
		key := strings.Join([]string{vsPrefix, "slot1"}, ".")
		ms.On(createCompositeKey, vsPrefix, []string{"slot1"}).Return(key, nil)
		ms.On(getState, key).Return([]byte{}, nil)
		ms.On(putState, key, anyBytes).Return(nil)
	}
	{
		key := strings.Join([]string{balancePrefix, patient164, "slot1"}, ".")
		ms.On(createCompositeKey, balancePrefix, []string{patient164, "slot1"}).Return(key, nil)
		ms.On(putState, key, anyBytes).Return(nil)
	}
	ms.On(setEvent, "Transfer", anyBytes).Return(nil)

	mci := &MockClientIdentity{}
	mci.On(getMSPID).Return("MedicalStationMSP", nil)

	mc := &MockContext{}
	mc.On(getStub).Return(ms)
	mc.On(getClientIdentity).Return(mci)

	return mc, ms, gen
}

func setupTestIssueSlot2() (*MockContext, *MockStub, TokenIdGeneratorInterface) {
	ms := &MockStub{}
	gen := &MockTokenIdGenerator{
		[]string{"slot1", "slot2"},
	}

	mci := &MockClientIdentity{}
	mci.On(getMSPID).Return("SomethingWrong", nil)

	mc := &MockContext{}
	mc.On(getStub).Return(ms)
	mc.On(getClientIdentity).Return(mci)

	return mc, ms, gen
}

func setupTestIssueSlot3() (*MockContext, *MockStub, TokenIdGeneratorInterface) {
	ms := &MockStub{}
	gen := &MockTokenIdGenerator{
		[]string{slot1, slot2},
	}

	anyBytes := mock.AnythingOfType("[]uint8")

	vs := &VaccinationSlot{
		VaccinationSlotData: VaccinationSlotData{
			Type: Delta,
			Date: func() VaccinationDate {
				val, err := time.Parse("2006-01-02", "2000-01-01")
				if err != nil {
					log.Fatal(err)
				}
				return VaccinationDate(val)
			}(),
		},
		TokenId: slot1,
		Owner:   patient1,
	}

	vsb, _ := json.Marshal(vs)

	patient164 := base64.StdEncoding.EncodeToString([]byte(patient1))

	{
		key := strings.Join([]string{balancePrefix, patient164, slot1}, ".")
		patient1Balance := &MockIterator{
			queries: []queryresult.KV{
				{
					Key:   key,
					Value: []byte(slot1),
				},
			},
		}
		ms.On(getStateByPartialCompositeKey, balancePrefix, []string{patient164}).Return(patient1Balance, nil)
	}
	{
		key := strings.Join([]string{vsPrefix, "slot1"}, ".")
		ms.On(createCompositeKey, vsPrefix, []string{"slot1"}).Return(key, nil)
		ms.On(getState, key).Return(vsb, nil)
		ms.On(putState, key, anyBytes).Return(nil)
	}
	{
		key := strings.Join([]string{balancePrefix, patient164, "slot1"}, ".")
		ms.On(createCompositeKey, balancePrefix, []string{patient164, "slot1"}).Return(key, nil)
		ms.On(putState, key, anyBytes).Return(nil)
	}
	ms.On(setEvent, "Transfer", anyBytes).Return(nil)

	mci := &MockClientIdentity{}
	mci.On(getMSPID).Return("MedicalStationMSP", nil)

	mc := &MockContext{}
	mc.On(getStub).Return(ms)
	mc.On(getClientIdentity).Return(mci)

	return mc, ms, gen
}

//</editor-fold>

//<editor-fold desc="Test MakeOffer">
func TestMakeOffer(t *testing.T) {
	t.Run("Correct", func(t *testing.T) {
		ctx, _, gen := setupTestMakeOffer1(patient1)
		c := &VaccinationContract{
			IdGenerator: gen,
		}
		offer, err := c.MakeOffer(ctx, slot1, patient2, slot2)
		assert.Nil(t, err)
		assert.NotEmpty(t, offer)
	})
	t.Run("Wrong recipient", func(t *testing.T) {
		ctx, _, gen := setupTestMakeOffer1(patient1)
		c := &VaccinationContract{
			IdGenerator: gen,
		}
		_, err := c.MakeOffer(ctx, slot1, patient2, slot1)
		assert.Error(t, err)
	})
	t.Run("Wrong sender", func(t *testing.T) {
		ctx, _, gen := setupTestMakeOffer1(patient1)
		c := &VaccinationContract{
			IdGenerator: gen,
		}
		_, err := c.MakeOffer(ctx, slot2, patient2, slot2)
		assert.Error(t, err)
	})

}

func setupTestMakeOffer1(patient string) (*MockContext, *MockStub, TokenIdGeneratorInterface) {
	ms := &MockStub{}

	gen := &MockTokenIdGenerator{
		[]string{offer1},
	}

	anyBytes := mock.AnythingOfType("[]uint8")

	newDate := func(str string) VaccinationDate {
		value, err := time.Parse("2006-01-02", str)
		if err != nil {
			log.Fatal(err)
		}
		return VaccinationDate(value)
	}

	vs1 := &VaccinationSlot{
		VaccinationSlotData: VaccinationSlotData{
			Type: Alpha,
			Date: newDate("2000-01-01"),
		},
		TokenId: slot1,
		Owner:   patient,
	}

	vs2 := &VaccinationSlot{
		VaccinationSlotData: VaccinationSlotData{
			Type: Bravo,
			Date: newDate("2000-01-02"),
		},
		TokenId: slot2,
		Owner:   patient2,
	}

	vsb1, _ := json.Marshal(&vs1)
	vbs2, _ := json.Marshal(&vs2)

	patient164 := base64.StdEncoding.EncodeToString([]byte(patient1))
	patient264 := base64.StdEncoding.EncodeToString([]byte(patient2))

	{
		key := strings.Join([]string{vsPrefix, slot1}, ".")
		ms.On(createCompositeKey, vsPrefix, []string{slot1}).Return(key, nil)
		ms.On(getState, key).Return(vsb1, nil)
	}
	{
		key := strings.Join([]string{vsPrefix, slot2}, ".")
		ms.On(createCompositeKey, vsPrefix, []string{slot2}).Return(key, nil)
		ms.On(getState, key).Return(vbs2, nil)
	}
	{
		key := strings.Join([]string{offerPrefix, patient164, offer1}, ".")
		ms.On(createCompositeKey, offerPrefix, []string{patient164, offer1}).Return(key, nil)
		ms.On(putState, key, anyBytes).Return(nil)
	}
	{
		key := strings.Join([]string{offerPrefix, patient264, offer1}, ".")
		ms.On(createCompositeKey, offerPrefix, []string{patient264, offer1}).Return(key, nil)
		ms.On(putState, key, anyBytes).Return(nil)
	}
	{
		key := strings.Join([]string{offerPrefix, offer1}, ".")
		ms.On(createCompositeKey, offerPrefix, []string{offer1}).Return(key, nil)
		ms.On(putState, key, anyBytes).Return(nil)
	}

	patient64 := base64.StdEncoding.EncodeToString([]byte(patient))

	mci := &MockClientIdentity{}
	mci.On(getID).Return(patient64, nil)

	mc := &MockContext{}
	mc.On(getStub).Return(ms)
	mc.On(getClientIdentity).Return(mci)

	return mc, ms, gen
}

//<editor-fold>

//<editor-fold desc="Test AcceptOffer">
func TestAcceptOffer(t *testing.T) {
	newDate := func(str string) VaccinationDate {
		value, err := time.Parse("2006-01-02", str)
		if err != nil {
			log.Fatal(err)
		}
		return VaccinationDate(value)
	}
	vs1 := VaccinationSlot{
		VaccinationSlotData: VaccinationSlotData{
			Type: Alpha,
			Date: newDate("2050-02-01"),
		},
		TokenId: slot1,
		Owner:   patient1,
	}

	vs2 := VaccinationSlot{
		VaccinationSlotData: VaccinationSlotData{
			Type: Bravo,
			Date: newDate("2050-02-02"),
		},
		TokenId: slot2,
		Owner:   patient2,
	}
	t.Run("Correct", func(t *testing.T) {
		ctx, ms, gen := setupTestAcceptOffer1(vs1, vs2)
		c := &VaccinationContract{
			IdGenerator: gen,
		}
		err := c.AcceptOffer(ctx, offer1)
		assert.Nil(t, err)
		ms.AssertNumberOfCalls(t, setEvent, 2)
	})
	t.Run("Offer doesn't exists", func(t *testing.T) {
		ctx, ms, gen := setupTestAcceptOffer1(vs1, vs2)
		c := &VaccinationContract{
			IdGenerator: gen,
		}
		err := c.AcceptOffer(ctx, offer2)
		assert.Error(t, err)
		anyBytes := mock.AnythingOfType("[]uint8")
		ms.AssertNotCalled(t, setEvent, "Transfer", anyBytes)
	})
	t.Run("Past", func(t *testing.T) {
		vs1 := vs1
		vs1.Date = newDate("2000-04-25")
		ctx, ms, gen := setupTestAcceptOffer1(vs1, vs2)
		c := &VaccinationContract{
			IdGenerator: gen,
		}
		err := c.AcceptOffer(ctx, offer1)
		assert.Error(t, err)
		anyBytes := mock.AnythingOfType("[]uint8")
		ms.AssertNotCalled(t, setEvent, "Transfer", anyBytes)
	})
	t.Run("Burned", func(t *testing.T) {
		vs1 := vs1
		vs1.Burned = true
		ctx, ms, gen := setupTestAcceptOffer1(vs1, vs2)
		c := &VaccinationContract{
			IdGenerator: gen,
		}
		err := c.AcceptOffer(ctx, offer1)
		assert.Error(t, err)
		anyBytes := mock.AnythingOfType("[]uint8")
		ms.AssertNotCalled(t, setEvent, "Transfer", anyBytes)
	})
	t.Run("Deadline 1", func(t *testing.T) {
		vs1 := vs1
		vs1.Previous = slot3
		ctx, ms, gen := setupTestAcceptOffer1(vs1, vs2)
		c := &VaccinationContract{
			IdGenerator: gen,
		}
		err := c.AcceptOffer(ctx, offer1)
		assert.Nil(t, err)
		ms.AssertNumberOfCalls(t, setEvent, 2)
	})
	t.Run("Deadline 2", func(t *testing.T) {
		vs1 := vs1
		vs1.Previous = slot4
		ctx, ms, gen := setupTestAcceptOffer1(vs1, vs2)
		c := &VaccinationContract{
			IdGenerator: gen,
		}
		err := c.AcceptOffer(ctx, offer1)
		assert.Error(t, err)
		anyBytes := mock.AnythingOfType("[]uint8")
		ms.AssertNotCalled(t, setEvent, "Transfer", anyBytes)
	})
}

func setupTestAcceptOffer1(vs1 VaccinationSlot, vs2 VaccinationSlot) (*MockContext, *MockStub, TokenIdGeneratorInterface) {
	ms := &MockStub{}

	gen := &MockTokenIdGenerator{
		[]string{offer1},
	}

	anyBytes := mock.AnythingOfType("[]uint8")

	vsb1, _ := json.Marshal(&vs1)
	vsb2, _ := json.Marshal(&vs2)

	newDate := func(str string) VaccinationDate {
		value, err := time.Parse("2006-01-02", str)
		if err != nil {
			log.Fatal(err)
		}
		return VaccinationDate(value)
	}

	vs3 := VaccinationSlot{
		VaccinationSlotData: VaccinationSlotData{
			Type: Alpha,
			Date: newDate("2050-01-15"),
		},
		TokenId: slot1,
		Owner:   patient1,
	}
	vs4 := VaccinationSlot{
		VaccinationSlotData: VaccinationSlotData{
			Type: Alpha,
			Date: newDate("2049-01-01"),
		},
		TokenId: slot1,
		Owner:   patient1,
	}

	vsb3, _ := json.Marshal(&vs3)
	vsb4, _ := json.Marshal(&vs4)

	patient164 := base64.StdEncoding.EncodeToString([]byte(patient1))
	patient264 := base64.StdEncoding.EncodeToString([]byte(patient2))

	{
		key := strings.Join([]string{vsPrefix, slot1}, ".")
		ms.On(createCompositeKey, vsPrefix, []string{slot1}).Return(key, nil)
		ms.On(getState, key).Return(vsb1, nil)
		ms.On(putState, key, anyBytes).Return(nil)
	}
	{
		key := strings.Join([]string{vsPrefix, slot2}, ".")
		ms.On(createCompositeKey, vsPrefix, []string{slot2}).Return(key, nil)
		ms.On(getState, key).Return(vsb2, nil)
		ms.On(putState, key, anyBytes).Return(nil)
	}
	{
		key := strings.Join([]string{vsPrefix, slot3}, ".")
		ms.On(createCompositeKey, vsPrefix, []string{slot3}).Return(key, nil)
		ms.On(getState, key).Return(vsb3, nil)
	}
	{
		key := strings.Join([]string{vsPrefix, slot4}, ".")
		ms.On(createCompositeKey, vsPrefix, []string{slot4}).Return(key, nil)
		ms.On(getState, key).Return(vsb4, nil)
	}
	{
		key := strings.Join([]string{balancePrefix, patient164, slot1}, ".")
		ms.On(createCompositeKey, balancePrefix, []string{patient164, slot1}).Return(key, nil)
		ms.On(delState, key).Return(nil)
	}
	{
		key := strings.Join([]string{balancePrefix, patient264, slot2}, ".")
		ms.On(createCompositeKey, balancePrefix, []string{patient264, slot2}).Return(key, nil)
		ms.On(delState, key).Return(nil)
	}
	{
		key := strings.Join([]string{balancePrefix, patient164, slot2}, ".")
		ms.On(createCompositeKey, balancePrefix, []string{patient164, slot2}).Return(key, nil)
		ms.On(putState, key, anyBytes).Return(nil)
	}
	{
		key := strings.Join([]string{balancePrefix, patient264, slot1}, ".")
		ms.On(createCompositeKey, balancePrefix, []string{patient264, slot1}).Return(key, nil)
		ms.On(putState, key, anyBytes).Return(nil)
	}
	{
		key := strings.Join([]string{offerPrefix, patient164, offer1}, ".")
		ms.On(createCompositeKey, offerPrefix, []string{patient164, offer1}).Return(key, nil)
		ms.On(delState, key).Return(nil)
	}
	{
		key := strings.Join([]string{offerPrefix, patient264, offer1}, ".")
		ms.On(createCompositeKey, offerPrefix, []string{patient264, offer1}).Return(key, nil)
		ms.On(delState, key).Return(nil)
	}
	{
		key := strings.Join([]string{offerPrefix, offer1}, ".")
		ms.On(createCompositeKey, offerPrefix, []string{offer1}).Return(key, nil)
		ms.On(delState, key).Return(nil)
		offer := &TradeOffer{
			Uuid:          offer1,
			Sender:        patient2,
			SenderItem:    slot2,
			Recipient:     patient1,
			RecipientItem: slot1,
		}
		offerBytes, _ := json.Marshal(offer)
		ms.On(getState, key).Return(offerBytes, nil)
	}
	{
		key := strings.Join([]string{offerPrefix, offer2}, ".")
		ms.On(createCompositeKey, offerPrefix, []string{offer2}).Return(key, nil)
		ms.On(getState, key).Return([]byte{}, fmt.Errorf("not found"))
	}
	{
		transfer := &Transfer{
			From:    patient1,
			To:      patient2,
			TokenId: slot1,
		}
		transferBytes, _ := json.Marshal(transfer)
		ms.On(setEvent, "Transfer", transferBytes).Return(nil)
	}
	{
		transfer := &Transfer{
			From:    patient2,
			To:      patient1,
			TokenId: slot2,
		}
		transferBytes, _ := json.Marshal(transfer)
		ms.On(setEvent, "Transfer", transferBytes).Return(nil)
	}

	mci := &MockClientIdentity{}
	mci.On(getID).Return(patient164, nil)

	mc := &MockContext{}
	mc.On(getStub).Return(ms)
	mc.On(getClientIdentity).Return(mci)

	return mc, ms, gen
}

//</editor-fold>

package offchain

//see https://github.com/hyperledger/fabric-samples/blob/master/asset-transfer-basic/chaincode-go/chaincode/smartcontract_test.go

import (
	"encoding/hex"
	"fmt"
	"os"
	"testing"

	"chaincode/offchain_rest/mocks"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-chaincode-go/shimtest"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-protos-go/msp"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

const cert = `-----BEGIN CERTIFICATE-----
MIICbjCCAhWgAwIBAgIQDOFK5ymReal7+p2habPWejAKBggqhkjOPQQDAjCBlTEQ
MA4GA1UEBhMHR2VybWFueTEPMA0GA1UECBMGQmVybGluMQ8wDQYDVQQHEwZCZXJs
aW4xFDASBgNVBAkTC0hhdXB0c3RyLiAxMQ4wDAYDVQQREwUxMDExNzEaMBgGA1UE
ChMRYXRlbC5ub2RlbmVjdC5jb20xHTAbBgNVBAMTFGNhLmF0ZWwubm9kZW5lY3Qu
Y29tMB4XDTE5MTAyMTEwMDUwMFoXDTI5MTAxODEwMDUwMFowgY0xEDAOBgNVBAYT
B0dlcm1hbnkxDzANBgNVBAgTBkJlcmxpbjEPMA0GA1UEBxMGQmVybGluMRQwEgYD
VQQJEwtIYXVwdHN0ci4gMTEOMAwGA1UEERMFMTAxMTcxDzANBgNVBAsTBmNsaWVu
dDEgMB4GA1UEAwwXQWRtaW5AYXRlbC5ub2RlbmVjdC5jb20wWTATBgcqhkjOPQIB
BggqhkjOPQMBBwNCAAQVvt/VE+1L+sIYQH0HklhrP/FXuryomsVGvWNMnvJUtqu+
8r5t8si56qApO41g2+WIJZrjUBYgdrSB2yRgQ2/8o00wSzAOBgNVHQ8BAf8EBAMC
B4AwDAYDVR0TAQH/BAIwADArBgNVHSMEJDAigCC1O2t3N76Q4z2wSagPevCdTjbv
RdCmMZops5IRJ8W4pTAKBggqhkjOPQQDAgNHADBEAiBx74S2GTEscgAKwmWL5RpD
y1cOxZNf4ydNmkTbfbB3yAIgPAoBX/zPDtWHRwrcXqnhGe/gRY0gH4kiiem3YFZE
6fM=
-----END CERTIFICATE-----
`

//go:generate counterfeiter -o mocks/chaincodestub.go -fake-name ChaincodeStub . chaincodeStub
type chaincodeStub interface {
	shim.ChaincodeStubInterface
}

//go:generate counterfeiter -o mocks/transaction.go -fake-name TransactionContext . transactionContext
type transactionContext interface {
	contractapi.TransactionContextInterface
}

func setCreator(t *testing.T, stub *mocks.ChaincodeStub, mspID string, idbytes []byte) {
	var err error
	sid := &msp.SerializedIdentity{Mspid: mspID, IdBytes: idbytes}
	b, err := proto.Marshal(sid)
	if err != nil {
		t.FailNow()
	}
	stub.GetCreatorReturns(b, err)
	require.NoError(t, err)
}

func TestPutData(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	setCreator(t, chaincodeStub, "org1MSP", []byte(cert))
	clientID, err := cid.New(chaincodeStub)
	require.NoError(t, err)
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	transactionContext.GetClientIdentityReturns(clientID)
	contract := RoamingSmartContract{}

	mspid, err := clientID.GetMSPID()
	require.NoError(t, err)
	os.Setenv("CORE_PEER_LOCALMSPID", mspid)
	//response, err := contract.SetSQLDBConn(transactionContext, "192.168.0.40", "3306", "nomad", "nomad", "private_db")

	// local store operation
	err = contract.StorePrivateData(transactionContext, "org2MSP", "data")
	require.NoError(t, err)

	// remote store operation
	os.Setenv("CORE_PEER_LOCALMSPID", "org2MSP")
	err = contract.StorePrivateData(transactionContext, "org2MSP", "data")
	require.NoError(t, err)

	// dangerous, wrong store operation on invalid peer
	os.Setenv("CORE_PEER_LOCALMSPID", "org3MSP")
	err = contract.StorePrivateData(transactionContext, "org2MSP", "data")
	require.Error(t, err)

	//require.Equal(t, "OK", response.Status, "Status unexpected")
}

type ledgerEntry []byte

var ledger map[string][]ledgerEntry = make(map[string][]ledgerEntry)

func dumpLedger() {
	for key, val := range ledger {
		fmt.Printf("LEDGER[%q] = ", key)
		for _, entry := range val {
			fmt.Printf("[%q], ", string(entry))
		}
		fmt.Printf("\n")
	}
}

func myPutState(arg1 string, arg2 []byte) error {
	log.Infof("WRITE ledger[%s] = %s\n", arg1, string(arg2))

	// insert into ledger, append to state
	value, ok := ledger[arg1]
	if !ok {
		ledger[arg1] = make([]ledgerEntry, 0)
	}
	// add data
	ledger[arg1] = append(value, arg2)

	dumpLedger()

	return nil
}

/*
func myGetState(arg1 string) ([]byte, error) {
	log.Infof("READ ledger[%s]\n", arg1)

	// query data from store
	value, ok := ledger[arg1]
	if !ok {
		return error
		ledger[arg1] = make([]ledgerEntry, 0)
	}
	// add data
	ledger[arg1] = append(value, arg2)

	dumpLedger()

	return nil
}*/

func shimSetCreator(t *testing.T, stub *shimtest.MockStub, mspID string, idbytes []byte) {
	sid := &msp.SerializedIdentity{Mspid: mspID, IdBytes: idbytes}
	b, err := proto.Marshal(sid)
	if err != nil {
		t.FailNow()
	}
	stub.Creator = b
}
func TestStoreSignature(t *testing.T) {
	shimStub := shimtest.NewMockStub("Test", nil)

	//clientID, err := cid.New(mockStub)
	//require.NoError(t, err)
	//transactionContext := &mocks.TransactionContext{}

	// test put and get state:
	shimStub.MockTransactionStart("init")
	shimStub.PutState("test", []byte("test"))
	shimStub.MockTransactionEnd("init")

	res, err := shimStub.GetState("test")
	require.NoError(t, err)
	if err != nil {
		log.Errorf("ERROR")
	}
	log.Infof("READ [%q]\n", hex.EncodeToString(res))

	shimSetCreator(t, shimStub, "org1MSP", []byte(cert))

	_, err = cid.New(shimStub)
	require.NoError(t, err)

	/*chaincodeStub := &mocks.ChaincodeStub{}
	setCreator(t, chaincodeStub, "org1MSP", []byte(cert))

	clientID, err := cid.New(chaincodeStub)
	require.NoError(t, err)
	transactionContext := &mocks.TransactionContext{}

	// tell the mock setup what to return
	transactionContext.GetStubReturns(mockStub)
	transactionContext.GetClientIdentityReturns(clientID)

	// tell the mock which functions to use
	chaincodeStub.CreateCompositeKeyCalls(shim.CreateCompositeKey)
	chaincodeStub.PutStateCalls(myPutState)

	contract := RoamingSmartContract{}

	mspid, err := clientID.GetMSPID()
	require.NoError(t, err)
	os.Setenv("CORE_PEER_LOCALMSPID", mspid)
	//response, err := contract.SetSQLDBConn(transactionContext, "192.168.0.40", "3306", "nomad", "nomad", "private_db")

	res, err := transactionContext.GetStub().CreateCompositeKey("name~id", []string{"me", "123"})
	fmt.Printf("got %s\n", hex.EncodeToString([]byte(res)))

	//chaincodeStub.PutState(res, []byte("\x1234"))
	//bb, err := chaincodeStub.GetState(res)
	//fmt.Printf("readback %s\n", hex.EncodeToString(bb))

	// local store operation
	key := "0x01234KEY"
	err = contract.StoreSignature(transactionContext, key, "SHA3", []byte("\x1234"))
	require.NoError(t, err)

	//dumpAllStates(t, chaincodeStub)
	//dumpCompositeKey(t, chaincodeStub, "owner~type~key")

	//checkState(t, chaincodeStub, "owner~type~key", "123")

	//require.Equal(t, "OK", response.Status, "Status unexpected")
	*/
}

func dumpAllStates(t *testing.T, stub *mocks.ChaincodeStub) {
	keysIter, err := stub.GetStateByRange("\x00", "z")
	require.NoError(t, err)
	if keysIter == nil {
		fmt.Printf("> no results found")
		return
	}

	defer keysIter.Close()

	for keysIter.HasNext() {
		resp, iterErr := keysIter.Next()
		require.NoError(t, iterErr)
		require.NotNil(t, resp)
		fmt.Printf("ledger[%s] = %s\n", resp.Key, resp.Value)
	}
}

func dumpCompositeKey(t *testing.T, stub *mocks.ChaincodeStub, compositeKey string) {
	fmt.Printf("> dumping composite key '%s'\n", compositeKey)

	myIterator, err := stub.GetStateByPartialCompositeKey(compositeKey, []string{"org1MSP"})
	require.NoError(t, err)
	require.NotNil(t, myIterator)
	defer myIterator.Close()

	for myIterator.HasNext() {
		resp, err := myIterator.Next()
		require.NoError(t, err)

		fmt.Printf("ledger[%s] = %s\n", resp.Key, resp.Value)
	}
}

// based on ideas from https://medium.com/coinmonks/tutorial-on-hyperledger-fabrics-chaincode-testing-44c3f260cb2b
func checkState(t *testing.T, stub *mocks.ChaincodeStub, name string, value string) {
	bytes, err := stub.GetState(name)
	require.NoError(t, err)
	require.NotNil(t, bytes)
	require.EqualValues(t, string(bytes), value)
}

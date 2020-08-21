package offchain

//see https://github.com/hyperledger/fabric-samples/blob/master/asset-transfer-basic/chaincode-go/chaincode/smartcontract_test.go

import (
	"encoding/hex"
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

func setCreator(t *testing.T, stub *shimtest.MockStub, mspID string, idbytes []byte) {
	sid := &msp.SerializedIdentity{Mspid: mspID, IdBytes: idbytes}
	b, err := proto.Marshal(sid)
	if err != nil {
		t.FailNow()
	}
	stub.Creator = b
}

func TestStoreSignature(t *testing.T) {
	contract := RoamingSmartContract{}
	shimStub := shimtest.NewMockStub("Test", nil)

	setCreator(t, shimStub, "org1MSP", []byte(cert))

	clientID, err := cid.New(shimStub)
	require.NoError(t, err)
	transactionContext := &mocks.TransactionContext{}

	transactionContext.GetStubReturns(shimStub)
	transactionContext.GetClientIdentityReturns(clientID)

	// start tx
	shimStub.MockTransactionStart("txid_dummy_init")

	// test storesignature
	key := "0x01234KEY"
	err = contract.StoreSignature(transactionContext, key, "SHA3", []byte("\x1234"))
	require.NoError(t, err)

	// execute tx
	shimStub.MockTransactionEnd("txid_dummy_init")

	dumpAllPartialStates(t, shimStub, "owner~type~key")

	/*chaincodeStub := &mocks.ChaincodeStub{}

	// tell the mock setup what to return

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

func TestPutAndGetState(t *testing.T) {
	shimStub := shimtest.NewMockStub("Test", nil)

	testData := []byte("test")

	// test put and get state:
	shimStub.MockTransactionStart("txid_dummy_init")
	shimStub.PutState("test", testData)
	shimStub.MockTransactionEnd("txid_dummy_init")

	res, err := shimStub.GetState("test")
	require.NoError(t, err)
	if err != nil {
		log.Errorf("ERROR")
	}
	log.Infof("READ [%s]\n", hex.EncodeToString(res))
	require.Equal(t, res, testData, "data mismatch")

	dumpAllStates(t, shimStub)
}

func dumpAllPartialStates(t *testing.T, stub *shimtest.MockStub, keydef string) {
	keysIter, err := stub.GetStateByPartialCompositeKey(keydef, []string{})
	require.NoError(t, err)
	if keysIter == nil {
		log.Infof("no results found")
		return
	}

	log.Infof("================= dumping ledger for partial key [%s.*]", keydef)

	defer keysIter.Close()

	for keysIter.HasNext() {
		resp, iterErr := keysIter.Next()
		require.NoError(t, iterErr)
		require.NotNil(t, resp)
		log.Infof("                  ledger[%s] = %s\n", resp.Key, resp.Value)
	}
	log.Infof("================= done")
}

func dumpAllStates(t *testing.T, stub *shimtest.MockStub) {
	keysIter, err := stub.GetStateByRange("", "")
	require.NoError(t, err)
	if keysIter == nil {
		log.Infof("no results found")
		return
	}

	log.Infof("================= dumping ledger")

	defer keysIter.Close()

	for keysIter.HasNext() {
		resp, iterErr := keysIter.Next()
		require.NoError(t, iterErr)
		require.NotNil(t, resp)
		log.Infof("                  ledger[%s] = %s\n", resp.Key, resp.Value)
	}
	log.Infof("================= done")
}

// based on ideas from https://medium.com/coinmonks/tutorial-on-hyperledger-fabrics-chaincode-testing-44c3f260cb2b
func checkState(t *testing.T, stub *shimtest.MockStub, name string, value string) {
	bytes, err := stub.GetState(name)
	require.NoError(t, err)
	require.NotNil(t, bytes)
	require.EqualValues(t, string(bytes), value)
}

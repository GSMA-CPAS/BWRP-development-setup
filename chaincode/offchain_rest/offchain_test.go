package offchain

import (
	"os"
	"testing"

	"chaincode/offchain_rest/mocks"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-protos-go/msp"
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
	contract := SmartContract{}

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

func TestStoreSignature(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	setCreator(t, chaincodeStub, "org1MSP", []byte(cert))
	clientID, err := cid.New(chaincodeStub)
	require.NoError(t, err)
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	transactionContext.GetClientIdentityReturns(clientID)
	contract := SmartContract{}

	mspid, err := clientID.GetMSPID()
	require.NoError(t, err)
	os.Setenv("CORE_PEER_LOCALMSPID", mspid)
	//response, err := contract.SetSQLDBConn(transactionContext, "192.168.0.40", "3306", "nomad", "nomad", "private_db")

	// local store operation
	err = contract.StoreSignature(transactionContext, "0x01234KEY", "SHA3", []byte("\x1234"))
	require.NoError(t, err)

	//require.Equal(t, "OK", response.Status, "Status unexpected")
}

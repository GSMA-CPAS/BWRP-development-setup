package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
	"github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric-samples/chaincode/offchain/go/mocks"
	"github.com/stretchr/testify/require"
)

const certOrg1 = `-----BEGIN CERTIFICATE-----
MIICKDCCAc+gAwIBAgIQL+17U1jds+R0gXXt0aONhjAKBggqhkjOPQQDAjBzMQsw
CQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZvcm5pYTEWMBQGA1UEBxMNU2FuIEZy
YW5jaXNjbzEZMBcGA1UEChMQb3JnMS5leGFtcGxlLmNvbTEcMBoGA1UEAxMTY2Eu
b3JnMS5leGFtcGxlLmNvbTAeFw0yMDA3MjcwODU3MDBaFw0zMDA3MjUwODU3MDBa
MGsxCzAJBgNVBAYTAlVTMRMwEQYDVQQIEwpDYWxpZm9ybmlhMRYwFAYDVQQHEw1T
YW4gRnJhbmNpc2NvMQ4wDAYDVQQLEwVhZG1pbjEfMB0GA1UEAwwWQWRtaW5Ab3Jn
MS5leGFtcGxlLmNvbTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABAQz/bUci79i
YyHYbKOajqTNO743eFClN2aByu+juLgYiCs20LCBAzwjs1TtdabaW8GGbyACeW2y
V6x+blFpAHijTTBLMA4GA1UdDwEB/wQEAwIHgDAMBgNVHRMBAf8EAjAAMCsGA1Ud
IwQkMCKAIJEHcS5Rmtj6A/rU8RNWWTv5wpfJ3npreKl3BQQ/Zd9dMAoGCCqGSM49
BAMCA0cAMEQCIBqKl4f3IhRWzdM9k2j7s8uKY6znlqkKfRDJv9bU/WmkAiAlUfHt
SgnzXlDiDhrSLtAGOFA8kp82cbiiK7KGnKrWog==
-----END CERTIFICATE-----
`

const certOrg2 = `-----BEGIN CERTIFICATE-----
MIICKTCCAc+gAwIBAgIQJ2q2ItMCY7m/FpxVGk0hCTAKBggqhkjOPQQDAjBzMQsw
CQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZvcm5pYTEWMBQGA1UEBxMNU2FuIEZy
YW5jaXNjbzEZMBcGA1UEChMQb3JnMi5leGFtcGxlLmNvbTEcMBoGA1UEAxMTY2Eu
b3JnMi5leGFtcGxlLmNvbTAeFw0yMDA3MjcwODU3MDBaFw0zMDA3MjUwODU3MDBa
MGsxCzAJBgNVBAYTAlVTMRMwEQYDVQQIEwpDYWxpZm9ybmlhMRYwFAYDVQQHEw1T
YW4gRnJhbmNpc2NvMQ4wDAYDVQQLEwVhZG1pbjEfMB0GA1UEAwwWQWRtaW5Ab3Jn
Mi5leGFtcGxlLmNvbTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABCID2c+kkNK7
888J1y/RA7GpT9PKqz6wG3Sz6jIaNiwPxwk7stxcKWqXcrz0mneKclbdg1G+3Wgb
hhioQuTJVYyjTTBLMA4GA1UdDwEB/wQEAwIHgDAMBgNVHRMBAf8EAjAAMCsGA1Ud
IwQkMCKAIAHnSIhfvIgXdibe/7i6u4a5vuPfaPmQ4+XmWaMbUMQFMAoGCCqGSM49
BAMCA0gAMEUCIQCHN86sr+R7Hbr4rFRw4jjJR1jF0sTOBSVS8HYSScS3UQIgZ3RO
bjNiqgfRr3guPXMzGUD479My93bh6DGo2FZofn0=
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

//go:generate counterfeiter -o mocks/historyqueryiterator.go -fake-name HistoryQueryIterator . historyQueryIterator
type historyQueryIterator interface {
	shim.HistoryQueryIteratorInterface
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

func setupChaincode(t *testing.T) (SmartContract, *mocks.ChaincodeStub, *mocks.TransactionContext) {
	chaincodeStub := &mocks.ChaincodeStub{}
	setCreator(t, chaincodeStub, "org1MSP", []byte(certOrg1))
	clientID, err := cid.New(chaincodeStub)
	require.NoError(t, err)

	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	transactionContext.GetClientIdentityReturns(clientID)

	contract := SmartContract{}
	err = contract.InitLedger(transactionContext)
	require.NoError(t, err)

	mspid, err := clientID.GetMSPID()
	require.NoError(t, err)
	os.Setenv("CORE_PEER_LOCALMSPID", mspid)
	response, err := contract.SetSQLDBConn(transactionContext, "127.0.0.1", "3306", "nomad", "Fe3gtZ6!s4Fe", "dtag")
	require.NoError(t, err)
	require.Equal(t, "OK", response.Status, "Status unexpected")

	response, err = contract.GetSQLDBConn(transactionContext)
	require.NoError(t, err)
	fmt.Println(response.Info)

	response, err = contract.Test(transactionContext)
	require.NoError(t, err)
	require.Equal(t, "OK", response.Status, "Status unexpected")

	return contract, chaincodeStub, transactionContext
}

func TestPutState(t *testing.T) {
	contract, _, transactionContext := setupChaincode(t)
	response, err := contract.PutState(transactionContext, "tx1", "val", "")
	require.NoError(t, err)
	require.Equal(t, "OK", response.Status, "Status unexpected")
}

func TestPutData(t *testing.T) {
	contract, chaincodeStub, transactionContext := setupChaincode(t)
	type Content struct {
		Txh string `json:"TXH"`
		Val string `json:"val"`
	}
	setCreator(t, chaincodeStub, "Org2MSP", []byte(certOrg2))
	//Create test data, here 1kb
	value := make([]byte, 1024)
	rand.Read(value)

	payload := Content{"123", string(value)}
	jsonPayload, err := json.Marshal(payload)
	mac := hmac.New(sha256.New, []byte("Org2MSP"))
	mac.Write(jsonPayload)
	hash := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	iterator := &mocks.HistoryQueryIterator{}
	iterator.HasNextReturnsOnCall(0, true)
	iterator.HasNextReturnsOnCall(1, false)

	timestamp, _ := ptypes.TimestampProto(time.Now())
	iterator.NextReturns(&queryresult.KeyModification{
		TxId:      "123",
		Timestamp: timestamp,
		Value:     []byte(hash),
		IsDelete:  false,
	}, nil)
	chaincodeStub.GetHistoryForKeyReturns(iterator, err)

	response, err := contract.PutData(transactionContext, "testKey", string(jsonPayload))
	require.NoError(t, err)
	require.Equal(t, "OK", response.Status, "Status unexpected")
}

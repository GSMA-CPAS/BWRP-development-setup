package main

import (
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-chaincode-go/shimtest"
	"github.com/hyperledger/fabric-protos-go/msp"
	"github.com/magiconair/properties/assert"
	"testing"
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

func setCreator(t *testing.T, stub *shimtest.MockStub, mspID string, idbytes []byte) {
	sid := &msp.SerializedIdentity{Mspid: mspID, IdBytes: idbytes}
	b, err := proto.Marshal(sid)
	if err != nil {
		t.FailNow()
	}
	stub.Creator = b
}

func checkInit(t *testing.T, stub *shimtest.MockStub, args [][]byte) {
	res := stub.MockInit("1", args)
	if res.Status != shim.OK {
		fmt.Println("Init failed", string(res.Message))
		t.FailNow()
	}
}

func checkStoreSignature(t *testing.T, stub *shimtest.MockStub, key string, signature string, signatureAlgo string) {
	res := stub.MockInvoke("1", [][]byte{[]byte("storeSignature"), []byte(key), []byte(signature), []byte(signatureAlgo)})
	if res.Status != shim.OK {
		fmt.Println("Storing signature failed", string(res.Message))
		t.FailNow()
	}
}

func checkGetSignature(t *testing.T, stub *shimtest.MockStub, key string) []byte {
	res := stub.MockInvoke("1", [][]byte{[]byte("getSignatures"), []byte(key)})
	if res.Status != shim.OK {
		fmt.Println("Retrieving signature failed", string(res.Message))
		t.FailNow()
	}
	if res.Payload == nil {
		fmt.Println("No payload retrieved")
		t.FailNow()
	}
	return res.GetPayload()
}

func checkStoreEndpoint(t *testing.T, stub *shimtest.MockStub) {
	res := stub.MockInvoke("1", [][]byte{[]byte("createOrUpdateOrganization"), []byte("{\"companyName\":\"Telekom Deutschland\",\"storageEndpoint\":\"https://storage.dtag.poc.com/api/v1/storage/endpoint\"}")})
	if res.Status != shim.OK {
		fmt.Println("Storing endpoint failed", string(res.Message))
		t.FailNow()
	}
}

func checkGetOwnEndpoint(t *testing.T, stub *shimtest.MockStub) {
	res := stub.MockInvoke("1", [][]byte{[]byte("getOrganization")})
	if res.Status != shim.OK {
		fmt.Println("Retrieving own endpoint failed", string(res.Message))
		t.FailNow()
	}
	var organizationData OrganizationData
	err := json.Unmarshal(res.Payload, &organizationData)
	if err != nil {
		fmt.Println("Could not unmarshal organization data: ", err)
		t.FailNow()
	}
	sId := &msp.SerializedIdentity{}
	err = proto.Unmarshal(stub.Creator, sId)
	if err != nil {
		fmt.Println("Could not retrieve MSP ID")
		t.FailNow()
	}
	mspid := sId.GetMspid()
	assert.Equal(t, organizationData.Mspid, mspid, "Own organization not as expected")
}

func checkGetRemoteEndpoint(t *testing.T, stub *shimtest.MockStub, mspid string) {
	res := stub.MockInvoke("1", [][]byte{[]byte("getOrganization"), []byte(mspid)})
	if res.Status != shim.OK {
		fmt.Println("Retrieving own endpoint failed", string(res.Message))
		t.FailNow()
	}
	var organizationData OrganizationData
	err := json.Unmarshal(res.Payload, &organizationData)
	if err != nil {
		fmt.Println("Could not unmarshal organization data: ", err)
		t.FailNow()
	}
	assert.Equal(t, organizationData.Mspid, mspid, "Remote organization not as expected")
}


func TestStoreSignature(t *testing.T) {
	var err error

	scc := new(AtomicSetupChaincode)
	stub := shimtest.NewMockStub("abac", scc)

	setCreator(t, stub, "org1MSP", []byte(cert))

	checkInit(t, stub, [][]byte{})

	key := "42"
	signature := "304402207f694c7075058ef8b01a98e62563a6b5c01243fe59a9ff9095c874db3fb8916502203dd6eb451f0990e33e1e079b1ae7aa5baa59b7dedbd2ca4ed928cffe8f497e3c"
	signatureAlgo := "ecdsa-with-SHA256_secp256r1"
	checkStoreSignature(t, stub, key, signature, signatureAlgo)


	var signatures Signatures
	payload := checkGetSignature(t, stub, key)
	err = json.Unmarshal(payload, &signatures)
	if err != nil {
		fmt.Println("Unmarschalling signatures failed", err.Error())
		t.FailNow()
	}
	assert.Equal(t, signatures[0].DocumentDataSignature, signature, "Signature not equal")
	assert.Equal(t, signatures[0].DocumentDataSignatureAlgo, signatureAlgo, "Signature algorithm not equal")
}

func TestStoreEndpoint(t *testing.T) {
	scc := new(AtomicSetupChaincode)
	stub := shimtest.NewMockStub("abac", scc)

	setCreator(t, stub, "org1MSP", []byte(cert))

	checkInit(t, stub, [][]byte{})

	checkStoreEndpoint(t, stub)

	checkGetOwnEndpoint(t, stub)

	setCreator(t, stub, "org2MSP", []byte(cert))

	checkGetRemoteEndpoint(t, stub, "org1MSP")
}
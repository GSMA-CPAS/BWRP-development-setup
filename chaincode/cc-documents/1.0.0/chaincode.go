// ====CHAINCODE EXECUTION SAMPLES (CLI) ==================
//
// === Create organization and store endpoint ===
//peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C mychannel -n atomic -c '{"Args":["createOrUpdateOrganization","{\"companyName\":\"Telekom Deutschland\",\"storageEndpoint\":\"https:\/\/storage.dtag.poc.com\/api\/v1\/storage\/endpoint\"}"]}'
//
// === Get own organization details ===
//peer chaincode query -C mychannel -n atomic -c '{"Args":["getOrganization"]}'
//
// === Store signature ===
//peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C mychannel -n atomic -c '{"Args":["storeSignature","42","304402207f694c7075058ef8b01a98e62563a6b5c01243fe59a9ff9095c874db3fb8916502203dd6eb451f0990e33e1e079b1ae7aa5baa59b7dedbd2ca4ed928cffe8f497e3c","ecdsa-with-SHA256_secp256r1"]}'
//
// === Get signature ===
//peer chaincode query -C mychannel -n atomic -c '{"Args":["getSignatures","42"]}'

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/msp"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

type AtomicSetupChaincode struct{}

type Signatures []Signature

type Signature struct {
	DocumentDataSignature     string `json:"DocumentDataSignature"`
	DocumentDataSignatureAlgo string `json:"documentDataSignature"`
	Timestamp                 int64  `json:"timestamp"`
	Certificate               []byte `json:"certificate"`
}

type OrganizationDataInput struct {
	CompanyName     string `json:"companyName"`
	StorageEndpoint string `json:"storageEndpoint"`
}

type OrganizationData struct {
	ObjectType string                `json:"docType"`
	Mspid      string                `json:"mspid"`
	Data       OrganizationDataInput `json:"data"`
}

func main() {
	err := shim.Start(new(AtomicSetupChaincode))
	if err != nil {
		fmt.Printf("Error starting Atomic Setup chaincode: %s", err)
	}
}

func (t *AtomicSetupChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (t *AtomicSetupChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	switch function {
	case "storeSignature":
		return t.storeSignature(stub, args)
	case "getSignatures":
		return t.getSignatures(stub, args)
	case "createOrUpdateOrganization":
		return t.createOrUpdateOrganization(stub, args)
	case "getOrganization":
		return t.getOrganization(stub, args)
	default:
		//error
		fmt.Println("invoke did not find func: " + function)
		return shim.Error("Received unknown function invocation")
	}
}

func (t *AtomicSetupChaincode) storeSignature(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 3 {
		return shim.Error("3 arguments required")
	}
	if len(args[0]) == 0 {
		return shim.Error("Key must not be empty")
	}
	if len(args[1]) == 0 {
		return shim.Error("DocumentDataSignature must not be empty")
	}
	if len(args[2]) == 0 {
		return shim.Error("DocumentDataSignatureAlgo must not be empty")
	}
	key := args[0]
	timestamp, err := stub.GetTxTimestamp()
	if err != nil {
		return shim.Error("Could not retrieve Tx timestamp")
	}
	cert, err := cid.GetX509Certificate(stub)
	if err != nil {
		return shim.Error("Could not retrieve sender certificate")
	}
	signature := Signature{
		DocumentDataSignature:     args[1],
		DocumentDataSignatureAlgo: args[2],
		Timestamp:                 timestamp.Seconds,
		Certificate:               cert.Raw,
	}

	signaturesAsBytes, err := stub.GetState(key)
	if err != nil {
		return shim.Error("Failed to get Signature: " + err.Error())
	}

	var signatures Signatures

	if len(signaturesAsBytes) > 0 {
		err = json.Unmarshal(signaturesAsBytes, &signatures)
		if err != nil {
			return shim.Error("Failed to parse document data: " + err.Error())
		}
	}

	signatures = append(signatures, signature)

	signaturesAsBytes, err = json.Marshal(signatures)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(key, signaturesAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(signaturesAsBytes)
}

func (t *AtomicSetupChaincode) getSignatures(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	key := args[0]

	signatures, err := stub.GetState(key)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(signatures)
}

func (t *AtomicSetupChaincode) createOrUpdateOrganization(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1 (OrganizationDataInput)")
	}

	passedOrganizationData := args[0]

	var organizationDataInput OrganizationDataInput
	err = json.Unmarshal([]byte(passedOrganizationData), &organizationDataInput)
	if err != nil {
		return shim.Error("Failed to decode JSON of: " + passedOrganizationData)
	}
	if len(organizationDataInput.CompanyName) == 0 {
		return shim.Error("CompanyName must be a non-empty string")
	}
	if len(organizationDataInput.StorageEndpoint) == 0 {
		return shim.Error("StorageEndpoint must be a non-empty string")
	}

	mspid, err := getIdentifier(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	organizationData := &OrganizationData{
		ObjectType: "organization",
		Mspid:      mspid,
		Data:       organizationDataInput,
	}

	organizationKey := "organization~" + mspid
	organizationDataBytes, err := json.Marshal(organizationData)

	err = stub.PutState(organizationKey, organizationDataBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *AtomicSetupChaincode) getOrganization(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	var mspid string
	if len(args) == 0 {
		mspid, err = getIdentifier(stub)
		if err != nil {
			return shim.Error(err.Error())
		}
	} else if len(args) == 1 {
		mspid = args[0]
	} else {
		return shim.Error("Incorrect number of arguments. Expecting 0 for own organization or 1 (mspid)")
	}

	organizationKey := "organization~" + mspid
	organzationBytes, err := stub.GetState(organizationKey)
	if err != nil || len(organzationBytes) == 0 {
		return shim.Error("Organization does not exist: " + mspid)
	}

	return shim.Success(organzationBytes)
}

func getIdentifier(stub shim.ChaincodeStubInterface) (string, error) {
	// GetCreator returns marshaled serialized identity of the client
	serializedID, _ := stub.GetCreator()

	sId := &msp.SerializedIdentity{}
	err := proto.Unmarshal(serializedID, sId)
	if err != nil {
		return "", errors.New("Could not retrieve MSP ID")
	}
	mspid := sId.GetMspid()
	return mspid, nil
}

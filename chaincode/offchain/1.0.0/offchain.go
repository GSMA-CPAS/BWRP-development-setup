/*
 */

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"

	//        "github.com/leesper/couchdb-golang"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type SmartContract struct {
	contractapi.Contract
}

type StatusResponse struct {
	Status string `json:"status"`
	Info   string `json:"info"`
}

type Transaction struct {
	ID        int    `json:"_id"`
	Timestamp int    `json:"_timestamp"`
	OrgRef    string `json:"org_ref"`
	Creator   string `json:"creator"`
	Key       string `json:"key"`
	Value     string `json:"value"`
	IsDeleted bool   `json:"_is_deleted"`
}

type Connection struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

var myConn Connection

// InitLedger not used. Just a place holder
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	fmt.Println(">InitLedger")
	return nil
}

// Test Function to make sure local DB is working.
// call by remote to make sure before we start the process.
func (s *SmartContract) Test(ctx contractapi.TransactionContextInterface) (StatusResponse, error) {
	fmt.Println(">Test")

	err := _checkExpiry(ctx)
	if err != nil {
		return StatusResponse{Status: "ERROR"}, err
	}

	db, err := connectMySQLDB()
	if err != nil {
		return StatusResponse{Status: "ERROR"}, err
	}
	db.Close()

	return StatusResponse{Status: "OK"}, nil
}

//PutState, Create Tranasction "Signature".
//To be invoked Creator locally.
func (s *SmartContract) PutState(ctx contractapi.TransactionContextInterface, key string, value string, old string) (StatusResponse, error) {
	fmt.Println(">PutState")

	err := _checkExpiry(ctx)
	if err != nil {
		return StatusResponse{Status: "ERROR"}, err
	}

	if len(old) > 0 {
		err := ctx.GetStub().DelState(old)
		if err != nil {
			return StatusResponse{Status: "ERROR"}, err
		}
	}

	err = ctx.GetStub().PutState(key, []byte(value))
	if err != nil {
		return StatusResponse{Status: "ERROR"}, err
	}

	return StatusResponse{Status: "OK", Info: "State Set"}, nil
}

//OffChain Push of Payload to Destination MSP
//To be invoke remotely from Creator
//ACL Control forbids to be invoked locally.
func (s *SmartContract) PutData(ctx contractapi.TransactionContextInterface, key string, data string) (StatusResponse, error) {
	fmt.Println(">PutData")

	err := _checkExpiry(ctx)
	if err != nil {
		return StatusResponse{Status: "ERROR"}, err
	}

	cid, err := cid.New(ctx.GetStub())
	if err != nil {
		return StatusResponse{Status: "ERROR"}, err
	}
	mspid, err := cid.GetMSPID()
	/*if os.Getenv("CORE_PEER_LOCALMSPID") == mspid {
		return StatusResponse{Status: "ERROR"}, fmt.Errorf("You do not have permission to call this function")
	}*/

	var x map[string]interface{}
	err2 := json.Unmarshal([]byte(data), &x)
	if err2 != nil {
		return StatusResponse{Status: "ERROR"}, fmt.Errorf("Payload not JSON")
	}

	if x["TXH"] == nil {
		return StatusResponse{Status: "ERROR"}, fmt.Errorf("Payload do not contain key 'TXH'")
	}

	history, err := ctx.GetStub().GetHistoryForKey(x["TXH"].(string))
	if err != nil {
		return StatusResponse{Status: "ERROR"}, fmt.Errorf("Hash Signature not found!")
	}

	for history.HasNext() {
		transaction, _ := history.Next()
		if transaction.IsDelete == false {
			fmt.Println("TransactionID=" + transaction.TxId)
			fmt.Printf("Current Timestamp=%v ,Transaction Timestamp=%v [diff=%v]\n", time.Now().Unix(), transaction.Timestamp.Seconds, (time.Now().Unix() - transaction.Timestamp.Seconds))
			fmt.Println("Transaction Value=" + string(transaction.Value))

			if ValidMAC(data, transaction.Value, mspid) == true && transaction.TxId == x["TXH"].(string) && (time.Now().Unix()-transaction.Timestamp.Seconds) < 8 {

				//add code to write data to localDB
				db, err := connectMySQLDB()
				if err != nil {
					return StatusResponse{Status: "ERROR"}, err
				}

				err = insertLocalDB(db, mspid, mspid, key, data)
				if err != nil {
					return StatusResponse{Status: "ERROR"}, err
				}

				db.Close()

				return StatusResponse{Status: "OK", Info: "Data Created"}, nil
			}

		}
	}

	return StatusResponse{Status: "ERROR"}, fmt.Errorf("Cannot verify data.")
}

//Verify Payload
//to be invoked by orhet Chaincode as well.
func (s *SmartContract) VerifyData(ctx contractapi.TransactionContextInterface, data string, creator string) (StatusResponse, error) {
	fmt.Println(">VerifyData")

	err := _checkExpiry(ctx)
	if err != nil {
		return StatusResponse{Status: "ERROR"}, err
	}

	var x map[string]interface{}
	err = json.Unmarshal([]byte(data), &x)
	if err != nil {
		return StatusResponse{Status: "ERROR"}, fmt.Errorf("Payload not JSON")
	}

	if x["TXH"] == nil {
		return StatusResponse{Status: "ERROR"}, fmt.Errorf("Payload do not contain key 'TXH'")
	}
	if x["timestamp"] == nil {
		return StatusResponse{Status: "ERROR"}, fmt.Errorf("Payload do not contain key 'timestamp'")
	}

	history, err := ctx.GetStub().GetHistoryForKey(x["TXH"].(string))
	if err != nil {
		return StatusResponse{Status: "ERROR"}, fmt.Errorf("Hash Signature not found!")
	}

	var transaction *queryresult.KeyModification
	var count = 0

	for history.HasNext() {
		transaction, _ = history.Next()
		count++
		fmt.Println("TransactionID=" + transaction.TxId)
		fmt.Printf("Created Timestamp=%v ,Transaction Timestamp=%v [diff=%v]\n", x["timestamp"].(float64), transaction.Timestamp.Seconds, (int64(x["timestamp"].(float64)) - transaction.Timestamp.Seconds))
		fmt.Println("Transaction Value=" + string(transaction.Value))
		fmt.Println("Transaction is_delete=" + strconv.FormatBool(transaction.IsDelete))
		if transaction.IsDelete == true {
			return StatusResponse{Status: "NOK", Info: "Payload is NOT Valid."}, nil
		}
	}

	if ValidMAC(data, transaction.Value, creator) == true && transaction.TxId == x["TXH"].(string) && (int64(x["timestamp"].(float64))-transaction.Timestamp.Seconds) < 8 && count == 1 {
		return StatusResponse{Status: "OK", Info: "Valid Payload"}, nil
	}

	return StatusResponse{Status: "NOK", Info: "Payload is NOT Valid."}, nil
}

//Verify Remote Data is the same
func (s *SmartContract) VerifyRemote(ctx contractapi.TransactionContextInterface, key string, hash string) (StatusResponse, error) {
	fmt.Println(">VerifyRemote")

	err := _checkExpiry(ctx)
	if err != nil {
		return StatusResponse{Status: "ERROR"}, err
	}

	cid, err := cid.New(ctx.GetStub())
	if err != nil {
		return StatusResponse{Status: "ERROR"}, err
	}
	mspid, err := cid.GetMSPID()
	/*if os.Getenv("CORE_PEER_LOCALMSPID") == mspid {
		return StatusResponse{Status: "ERROR"}, errors.New("Cannot be called locally.")
	}*/

	db, err := connectMySQLDB()
	if err != nil {
		return StatusResponse{Status: "ERROR"}, err
	}

	local_trans, err := readLocalDB(db, mspid, key)
	if err != nil {
		return StatusResponse{Status: "ERROR"}, err
	}

	db.Close()

	var x map[string]interface{}
	err = json.Unmarshal([]byte(local_trans.Value), &x)
	if err != nil {
		return StatusResponse{Status: "ERROR"}, fmt.Errorf("Payload not JSON")
	}

	history, err := ctx.GetStub().GetHistoryForKey(x["TXH"].(string))
	if err != nil {
		return StatusResponse{Status: "ERROR"}, fmt.Errorf("Hash Signature not found!")
	}

	var transaction *queryresult.KeyModification
	var count = 0

	for history.HasNext() {
		transaction, _ = history.Next()
		count++
		fmt.Println("TransactionID=" + transaction.TxId)
		fmt.Println("Transaction Value=" + string(transaction.Value))
		fmt.Println("Transaction is_delete=" + strconv.FormatBool(transaction.IsDelete))
		if transaction.IsDelete == true {
			return StatusResponse{Status: "NOK", Info: "Payload is NOT Valid."}, nil
		}
	}

	if ValidMAC(local_trans.Value, transaction.Value, local_trans.Creator) == true && ValidMAC(local_trans.Value, []byte(hash), local_trans.Creator) == true && transaction.TxId == x["TXH"].(string) && count == 1 {
		return StatusResponse{Status: "OK", Info: "Remote Valid"}, nil
	}

	return StatusResponse{Status: "NOK", Info: "Remote Copy is not the same."}, nil
}

//Local Chaincode Administration.
//Return currently configured "SQLDB" connection
//ACL only allowed to be called locally on peer by an adminstrator
func (s *SmartContract) GetSQLDBConn(ctx contractapi.TransactionContextInterface) (StatusResponse, error) {
	fmt.Println(">GetSQLDBConn")

	err := ACL(ctx)
	if len(err) > 0 {
		return StatusResponse{Status: "ERROR"}, fmt.Errorf(err)
	}

	if len(myConn.Host) == 0 || len(myConn.Port) == 0 || len(myConn.User) == 0 || len(myConn.Password) == 0 || len(myConn.Database) == 0 {
		return StatusResponse{Status: "ERROR"}, fmt.Errorf("SQLDBConn not Set")
	} else {
		out, _ := json.Marshal(myConn)

		return StatusResponse{Status: "OK", Info: string(out)}, nil
	}
}

//Local Chaincode Administration.
//Set or Update "SQLDB" connection
//ACL only allowed to be called locally on peer by an adminstrator
func (s *SmartContract) SetSQLDBConn(ctx contractapi.TransactionContextInterface, host string, port string, user string, password string, database string) (StatusResponse, error) {
	fmt.Println(">SetSQLDBConn")

	err := ACL(ctx)
	if len(err) > 0 {
		return StatusResponse{Status: "ERROR"}, fmt.Errorf(err)
	}

	if len(host) == 0 {
		return StatusResponse{Status: "ERROR"}, errors.New("host is empty")
	}
	myConn.Host = host

	if len(port) == 0 {
		return StatusResponse{Status: "ERROR"}, errors.New("port is empty")
	}
	myConn.Port = port

	if len(user) == 0 {
		return StatusResponse{Status: "ERROR"}, errors.New("user is empty")
	}
	myConn.User = user

	if len(password) == 0 {
		return StatusResponse{Status: "ERROR"}, errors.New("password is empty")
	}
	myConn.Password = password

	if len(database) == 0 {
		return StatusResponse{Status: "ERROR"}, errors.New("database is empty")
	}
	myConn.Database = database

	out, _ := json.Marshal(myConn)

	return StatusResponse{Status: "OK", Info: string(out)}, nil
}

//Connect to MYSQLDB instance
func connectMySQLDB() (*sql.DB, error) {
	if len(myConn.Host) == 0 || len(myConn.Port) == 0 || len(myConn.User) == 0 || len(myConn.Password) == 0 || len(myConn.Database) == 0 {
		return nil, errors.New("SQLDBConn not Set")
	} else {
		db, err := sql.Open("mysql", myConn.User+":"+myConn.Password+"@tcp("+myConn.Host+":"+myConn.Port+")/"+myConn.Database)
		if err != nil {
			return nil, err
		}
		err = db.Ping()
		if err != nil {
			return nil, errors.New("LocalDB not responding.")
		}
		return db, nil
	}
}

//Create Transaction into LocalDB
func insertLocalDB(db *sql.DB, argv1 string, argv2 string, argv3 string, argv4 string) error {
	//mark all previous entry as "deleted" (override)
	_, err := db.Exec("UPDATE local_data set `_is_deleted` = true WHERE `org_ref` = ? AND `key`= ?", argv1, argv3)
	if err != nil {
		return err
	}

	//insert entry.
	_, err = db.Exec("INSERT INTO local_data(`org_ref`, `creator`, `key`, `value`) VALUES(?, ?, ?, ?)", argv1, argv2, argv3, argv4)
	if err != nil {
		return err
	}

	db.Close()
	return nil
}

//read MYSQL Row Example.
func readLocalDB(db *sql.DB, argv1 string, argv2 string) (Transaction, error) {
	var trans Transaction

	err := db.QueryRow("SELECT a.key, a.value, a.creator FROM local_data a WHERE a.org_ref = ? AND a.key = ? AND a._is_deleted = 0", argv1, argv2).Scan(&trans.Key, &trans.Value, &trans.Creator)
	if err != nil {
		return Transaction{}, err
	}

	db.Close()

	return trans, nil
}

//function to check cert validity.
func _checkExpiry(ctx contractapi.TransactionContextInterface) error {
	/*fmt.Println(">_checkExpiry")

	        cid, err := cid.New(ctx.GetStub())
	        if err != nil {
	                return err
	        }

	        x509, err := cid.GetX509Certificate()
	        issued := x509.NotBefore
	        expire := x509.NotAfter

	        //check cert length. not more than 1 year
	        //24hr * 365days = 8760
	        oneyear,_ := time.ParseDuration("8760h0m0s")
	        difference := expire.Sub(issued)
	        if (difference > oneyear) {
	                fmt.Println(">_checkExpiry: Cert issued length is longer than 1 year!")
	//                return errors.New("Cert issued length is longer than 1 year!")
	        }

	        if (issued.After(time.Now())) {
	                fmt.Println(">_checkExpiry: Cert issued date is after Now!")
	//                return errors.New("Cert issued date is after Now!")
	        }

	        beforeonemonth := time.Unix(expire.Unix() - (30 * 24 * 3600), 0)
	        if (beforeonemonth.Before(time.Now())) {
	                fmt.Println(">_checkExpiry: Cert is expiring in 30 days!")
	//                return errors.New("Cert is expiring in 30 days!")
	        }*/
	return nil
}

//ACL Control
//Only permit Local MSP and request to be admin
func ACL(ctx contractapi.TransactionContextInterface) string {
	/*fmt.Println(">ACL")
	cid, err := cid.New(ctx.GetStub())
	if err != nil {
		return err.Error()
	}
	mspid, err := cid.GetMSPID()
	if os.Getenv("CORE_PEER_LOCALMSPID") != mspid {
		return "You do not have permission to call this function"
	}

	found, err := cid.HasOUValue("admin")
	if found == false {
		return "You do not have permission to call this function"
	}*/
	return ""
}

//Calculate and verify HMAC_SHA256 key, message
func ValidMAC(message string, messageMAC_byte []byte, key string) bool {
	fmt.Println(">ValidMAC")
	key_byte := []byte(key)
	message_byte := []byte(message)

	mac := hmac.New(sha256.New, key_byte)
	mac.Write(message_byte)
	expectedMAC := mac.Sum(nil)
	fmt.Println(">Expected HMAC=" + string(messageMAC_byte) + ", Calculated HMAC=" + base64.StdEncoding.EncodeToString(expectedMAC))
	decoded, err := base64.StdEncoding.DecodeString(string(messageMAC_byte))
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	return hmac.Equal(decoded, expectedMAC)
}

func main() {
	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create offchain chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting offchain chaincode: %s", err.Error())
	}
}

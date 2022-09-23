package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

type liveStock struct {
	OrgCode             string `json:"plantCode"`
	AllowNo             string `json:"allowNo"`
	AllowDate           string `json:"allowDate"`
	Extend              int    `json:"extend"`
	CarLicense          string `json:"carLicense"`
	CarQueue            string `json:"carQueue"`
	QueueTime           string `json:"queueTime"`
	LineNo              int    `json:"lineNo"`
	ShiftNo             string `json:"shiftNo"`
	FarmCode            string `json:"farmCode"`
	FarmName            string `json:"farmName"`
	HouseCode           int    `json:"houseCode"`
	HouseName           string `json:"houseName"`
	FarmOrg             string `json:"farmOrg"`
	ProductCode         string `json:"productCode"`
	ProductName         int    `json:"productName"`
	AwgWeight           string `json:"awgWeight"`
	Quantity            string `json:"quantity"`
	FarmArrivalDatetime string `json:"farmArrivalDateTime"`
	CancelFlag          int    `json:"cancelFlag"`
	CreateDateTime      string `json:"createDateTime"`
	CurrentState        string `json:"currentState"`
	DocType             string `json:"docType"`
	UnixTimeStamp       string `json:"unixTimestamp"`
	JobInfor            int    `json:"jobInfor"`
	PodInfor            string `json:"podInfor"`
	CatchInfor          int    `json:"catchInfor"`
	FactoryInfor        string `json:"factoryInfor"`
}

const (
	AccountKey  = "Account-%s"
	SuccessFlag = "Success"
	FailFlag    = "Fail"
)

type LiveStockChainCode struct {
}

func (i *LiveStockChainCode) Invoke(stub shim.ChaincodeStubInterface) (res peer.Response) {

	defer func() {
		if r := recover(); r != nil {
			res = shim.Error(fmt.Sprintf("%v", r))
		}
	}()

	fcn, args := stub.GetFunctionAndParameters()

	switch fcn {
	case "init":
		res = i.init(stub, args)
	case "invoke":
		res = i.invoke(stub, args)
	case "query":
		res = i.query(stub, args)
	case "delete":
		res = i.delete(stub, args)
	default:
		res = shim.Error("invalid function name")
	}
	return
}

func (i *LiveStockChainCode) init(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return formatError("incorrect number of arguments, expecting 2")
	}

	return formatSuccess("successfully")
}

func (i *LiveStockChainCode) invoke(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	var liveStock liveStock

	if len(args) != 1 {
		return formatError("Incorrect number of argument")
	}
	err := json.Unmarshal([]byte(args[0]), &liveStock)
	if err != nil {
		return formatError(err.Error())
	}
	timestamp := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
	blockKeyx := liveStock.DocType + liveStock.OrgCode + "-" + timestamp + "-" + liveStock.AllowNo
	blockKey := blockKeyx
	ArgsAsbyte, err := stub.GetState(blockKey)
	if err != nil {
		return formatError("Failed to get" + err.Error())
	} else if ArgsAsbyte != nil {
		return formatError("This document already exists: " + blockKey)
	}

	assetJSON, err := json.Marshal(liveStock)
	if err != nil {
		return formatError(err.Error())
	}

	err = stub.PutState(blockKey, assetJSON)
	if err != nil {
		return formatError(err.Error())
	}

	return formatSuccess("successfully")
}

func (i *LiveStockChainCode) query(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	var err error

	if len(args) != 1 {
		return formatError("Incorrect number of arguments")
	}

	blockKey := args[0]
	valAsbytes, err := stub.GetState(blockKey)
	if err != nil {
		return formatError("Failed to get" + err.Error())
	} else if valAsbytes == nil {
		return formatError("Document does not exist: " + blockKey)
	}

	errJson := fmt.Sprintf("{\"isSuccess\":true,\"message\":%s}", string(valAsbytes))
	return shim.Success([]byte(errJson))
}

func (i *LiveStockChainCode) delete(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	var err error

	if len(args) != 1 {
		return formatError("Incorrect number of arguments")
	}

	blockKey := args[0]
	valAsbytes, err := stub.GetState(blockKey)
	if err != nil {
		return formatError("Failed to get" + err.Error())
	} else if valAsbytes == nil {
		return formatError("Document does not exist: " + blockKey)
	}

	err = stub.DelState(blockKey) //remove the marble from chaincode state
	if err != nil {
		return formatError("Failed to delete state:" + err.Error())
	}

	return formatSuccess("successfully")
}

func (c *LiveStockChainCode) QueryFarmTransactionWithPagination(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	//   0
	// "queryString"
	if len(args) < 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	queryString := args[0]
	//return type of ParseInt is int64
	pageSize, err := strconv.ParseInt(args[1], 10, 32)
	if err != nil {
		return shim.Error(err.Error())
	}
	bookmark := args[2]

	queryResults, err := getQueryResultForQueryStringWithPagination(stub, queryString, int32(pageSize), bookmark)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

func getQueryResultForQueryStringWithPagination(stub shim.ChaincodeStubInterface, queryString string, pageSize int32, bookmark string) ([]byte, error) {

	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, responseMetadata, err := stub.GetQueryResultWithPagination(queryString, pageSize, bookmark)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	buffer, err := constructQueryResponseFromIterator(resultsIterator)
	if err != nil {
		return nil, err
	}

	bufferWithPaginationInfo := addPaginationMetadataToQueryResults(buffer, responseMetadata)

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", bufferWithPaginationInfo.String())

	return buffer.Bytes(), nil
}

func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) (*bytes.Buffer, error) {
	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("{\"Record\":[")
	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("],")

	return &buffer, nil
}

func addPaginationMetadataToQueryResults(buffer *bytes.Buffer, responseMetadata *peer.QueryResponseMetadata) *bytes.Buffer {

	buffer.WriteString("\"ResponseMetadata\":{\"RecordsCount\":")
	buffer.WriteString("\"")
	buffer.WriteString(fmt.Sprintf("%v", responseMetadata.FetchedRecordsCount))
	buffer.WriteString("\"")
	buffer.WriteString(", \"Bookmark\":")
	buffer.WriteString("\"")
	buffer.WriteString(responseMetadata.Bookmark)
	buffer.WriteString("\"}}")

	return buffer
}

func formatError(message string) peer.Response {
	errJson := fmt.Sprintf("{\"isSuccess\":false,\"message\":\"%s\"}", message)
	return shim.Error(errJson)
}

func formatSuccess(message string) peer.Response {
	errJson := fmt.Sprintf("{\"isSuccess\":true,\"message\":\"%s\"}", message)
	return shim.Success([]byte(errJson))
}

func main() {
	err := shim.Start(new(LiveStockChainCode))
	if err != nil {
		fmt.Printf("Error: %s", err)
	}
}

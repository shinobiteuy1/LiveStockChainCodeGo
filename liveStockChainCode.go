package main

import (
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

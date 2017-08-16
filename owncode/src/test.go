/*
Copyright IBM Corp 2016 All Rights Reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
		 http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"errors"
	"fmt"
	"encoding/json"
	"strconv"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type Container struct {
	Id string `json:"ContainerId"`
	Owner string `json:"Owner"`
}

type Account struct {
	Id string `json:"AccountId"`
	Balance float32	`json:"Balance"`
}

type Order struct {
	Id string `json:"OrderId"`
	Container string `json:"Container"`
	Customer string `json:"Customer"`
	Content string `json:"Content"`
	Destination string `json:"Destination"`
	Status string `json:"Status"`
	DefinedTransactions [3][2]string `json:"DefinedTransactionss"`
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init resets all the things
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	err := stub.PutState("ChaincodeOwner", []byte(args[0]))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// Invoke isur entry point to invoke a chaincode function
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "SetAsset" {
		return t.SetAsset(stub, args)
	} else if function == "ContainerHistorian" {
		return t.write(stub, args)
	}
	fmt.Println("invoke did not find func: " + function)

	return nil, errors.New("Received unknown function invocation: " + function)
}

// Query is our entry point for queries
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read" { //read a variable
		return Read(stub, args)
	}
	fmt.Println("query did not find func: " + function)

	return nil, errors.New("Received unknown function query: " + function)
}

func (t *SimpleChaincode) SetAsset(stub shim.ChaincodeStubInterface, args []string) ([]byte, error){
	
	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}
		
	assettype := args[0]
	if assettype == "Order" {
		var neworder Order
		err := json.Unmarshal([]byte(args[1]), &neworder)
		if err != nil {
			return nil, errors.New("Your Order seems to have incorrect parameters")
		}
		assetAsBytes, _ := json.Marshal(neworder)                      
		err = stub.PutState(neworder.Id, assetAsBytes)
		if err != nil {
			return nil, errors.New("Unable to place Order.")
		}
	} else if assettype == "Account" {
		var newacc Account
		err := json.Unmarshal([]byte(args[1]), &newacc)
		if err != nil {
			return nil, errors.New("Your Account seems to have incorrect parameters")
		}
		assetAsBytes, _ := json.Marshal(newacc)                      
		err = stub.PutState(newacc.Id, assetAsBytes)
		if err != nil {
			return nil, errors.New("Unable to create Account.")
		}
	} else if assettype == "Container" {
		var newcont Container
		err := json.Unmarshal([]byte(args[1]), &newcont)
		if err != nil {
			return nil, errors.New("Your Container seems to have incorrect parameters")
		}
		assetAsBytes, _ := json.Marshal(newcont)                      
		err = stub.PutState(newcont.Id, assetAsBytes)
		if err != nil {
			return nil, errors.New("Unable to create Container.")
		}
	}
	
	return []byte("A new asset was created!"), nil
}

func (t *SimpleChaincode) UpdateOrderStatus(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. Oder ID and Status")
	}

	orderid := args[0]
	orderstatus := args[1]
	
	orderAsBytes, err := stub.GetState(orderid)
	if err != nil {
		return nil, errors.New("Failed to get Order:" + err.Error())
	} else if orderAsBytes == nil {
		return nil, errors.New("Order does not exist")
	}

	orderToUpdate := Order{}
	err = json.Unmarshal(orderAsBytes, &orderToUpdate)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	
	//transaction execution
		//get customer account
	customeraccountAsBytes, err := stub.GetState(orderToUpdate.Customer)
	if err != nil {
		return nil, errors.New("Failed to get Customer Account:" + err.Error())
	} else if customeraccountAsBytes == nil {
		return nil, errors.New("Customer Account does not exist")
	}
	
	customerAccount := Account{}
	err = json.Unmarshal(customeraccountAsBytes, &customerAccount)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	
		//get operator account id and payment amount
	i := 0
	if orderstatus == "Beacon1" {
		i = 0
	} else if orderstatus == "Beacon2" {
		i = 1
	} else if orderstatus == "Beacon3" {
		i = 2
	} else {
		return nil, errors.New("Wrong status. Possible are Beacon1, Beacon2 and Beacon3")
	}			
	operatorAccountId := orderToUpdate.DefinedTransactions[i][0]
	auxvalue, err := strconv.ParseFloat(orderToUpdate.DefinedTransactions[i][1], 32)
	paymentAmount := float32(auxvalue)

		//get operator account
	operatoraccountAsBytes, err := stub.GetState(operatorAccountId)
	if err != nil {
		return nil, errors.New("Failed to get Operator Account:" + err.Error())
	} else if operatoraccountAsBytes == nil {
		return nil, errors.New("Opewrator Account does not exist")
	}
	
	operatorAccount := Account{}
	err = json.Unmarshal(operatoraccountAsBytes, &operatorAccount)
	if err != nil {
		return nil, errors.New(err.Error())
	}	
	
		//calculate new balances
	customerAccount.Balance = customerAccount.Balance - paymentAmount
	operatorAccount.Balance = operatorAccount.Balance + paymentAmount


	//rewrite customerAccount
	customerAccountUpdate, _ := json.Marshal(customerAccount)
	err = stub.PutState(customerAccount.Id, customerAccountUpdate)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	
	//rewrite operatorAccount
	operatorAccountUpdate, _ := json.Marshal(operatorAccount)
	err = stub.PutState(operatorAccount.Id, operatorAccountUpdate)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	
	//rewrite order
	orderToUpdate.Status = orderstatus
	orderJSONToUpdate, _ := json.Marshal(orderToUpdate)
	err = stub.PutState(orderid, orderJSONToUpdate)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	return []byte("Status changed and transaction executed."), nil
	
}


// write - invoke function to write key/value pair
func (t *SimpleChaincode) write(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var key, value string
	var err error
	fmt.Println("running write()")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the key and value to set")
	}

	key = args[0] //rename for funsies
	value = args[1]
	err = stub.PutState(key, []byte(value)) //write the variable into the chaincode state
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// read - query function to read key/value pair
func Read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var key, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the key to query")
	}

	key = args[0]
	valAsbytes, err := stub.GetState(key)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + key + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil
}
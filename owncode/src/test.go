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
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type Container struct {
	Id string
	Owner string
}

type Account struct {
	Id string
	Balance float32	
}

type Order struct {
	Id string
	Container string 
	Customer string
	Content string
	Destination string
	Status string
	DefinedTransactions map[string] map[string] float32
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
		var newasset Order
		err := json.Unmarshal([]byte(args[1]), &newasset)
		if err != nil {
			return nil, errors.New("Your Order seems to have incorrect parameters")
		}
		assetAsBytes, _ := json.Marshal(newasset)                      
		err = stub.PutState(newasset.Id, assetAsBytes)
		if err != nil {
			return nil, errors.New("Unable to place Order.")
		}
	} else if assettype == "Account" {
		var newasset Account
		err := json.Unmarshal([]byte(args[1]), &newasset)
		if err != nil {
			return nil, errors.New("Your Account seems to have incorrect parameters")
		}
		assetAsBytes, _ := json.Marshal(newasset)                      
		err = stub.PutState(newasset.Id, assetAsBytes)
		if err != nil {
			return nil, errors.New("Unable to create Account.")
		}
	} else if assettype == "Container" {
		var newasset Container
		err := json.Unmarshal([]byte(args[1]), &newasset)
		if err != nil {
			return nil, errors.New("Your Container seems to have incorrect parameters")
		}
		assetAsBytes, _ := json.Marshal(newasset)                      
		err = stub.PutState(newasset.Id, assetAsBytes)
		if err != nil {
			return nil, errors.New("Unable to create Container.")
		}
	}
	
	return []byte("A new Order was placed!"), nil
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
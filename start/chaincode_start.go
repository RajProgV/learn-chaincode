/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at
  http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package main

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	//"https://github.com/test56tester28tt/fabric/core/chaincode/shim"

	//===========addded start================
	"encoding/json"
	//"strconv"
	"strings"
	"time"
	//===========addded end================
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

//============start==========added globle var===============
var cpPrefix = "cp:"
var accountPrefix = "acct:"

//============end==========added globle var===============

//===========start======added for account creation ================
func generateCUSIPSuffix(issueDate string, days int) (string, error) {

	t, err := msToTime(issueDate)
	if err != nil {
		return "", err
	}

	maturityDate := t.AddDate(0, 0, days)
	month := int(maturityDate.Month())
	day := maturityDate.Day()

	suffix := seventhDigit[month] + eigthDigit[day]
	return suffix, nil

}

const (
	millisPerSecond     = int64(time.Second / time.Millisecond)
	nanosPerMillisecond = int64(time.Millisecond / time.Nanosecond)
)

func msToTime(ms string) (time.Time, error) {
	msInt, err := strconv.ParseInt(ms, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(msInt/millisPerSecond,
		(msInt%millisPerSecond)*nanosPerMillisecond), nil
}

type Owner struct {
	Company  string `json:"company"`
	Quantity int    `json:"quantity"`
}

type CP struct {
	CUSIP     string  `json:"cusip"`
	Ticker    string  `json:"ticker"`
	Par       float64 `json:"par"`
	Qty       int     `json:"qty"`
	Discount  float64 `json:"discount"`
	Maturity  int     `json:"maturity"`
	Owners    []Owner `json:"owner"`
	Issuer    string  `json:"issuer"`
	IssueDate string  `json:"issueDate"`
}

type Account struct {
	ID          string   `json:"id"`
	Prefix      string   `json:"prefix"`
	CashBalance float64  `json:"cashBalance"`
	AssetsIds   []string `json:"assetIds"`
}

type Transaction struct {
	CUSIP       string  `json:"cusip"`
	FromCompany string  `json:"fromCompany"`
	ToCompany   string  `json:"toCompany"`
	Quantity    int     `json:"quantity"`
	Discount    float64 `json:"discount"`
}

//===========end======added for account creation ================

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Printf("Init called, initializing chaincode")

	var A, B string    // Entities
	var Aval, Bval int // Asset holdings
	var err error

	if len(args) != 4 {
		return nil, errors.New("Incorrect number of arguments. Expecting 4")
	}

	// Initialize the chaincode
	A = args[0]
	Aval, err = strconv.Atoi(args[1])
	if err != nil {
		return nil, errors.New("Expecting integer value for asset holding")
	}
	B = args[2]
	Bval, err = strconv.Atoi(args[3])
	if err != nil {
		return nil, errors.New("Expecting integer value for asset holding")
	}
	fmt.Printf("Aval = %d, Bval = %d\n", Aval, Bval)

	// Write the state to the ledger
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return nil, err
	}

	err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// Transaction makes payment of X units from A to B
func (t *SimpleChaincode) invoke(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Printf("Running invoke")

	var A, B string    // Entities
	var Aval, Bval int // Asset holdings
	var X int          // Transaction value
	var err error

	if len(args) != 3 {
		return nil, errors.New("Incorrect number of arguments. Expecting 3")
	}

	A = args[0]
	B = args[1]

	// Get the state from the ledger
	// TODO: will be nice to have a GetAllState call to ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		return nil, errors.New("Failed to get state")
	}
	if Avalbytes == nil {
		return nil, errors.New("Entity not found")
	}
	Aval, _ = strconv.Atoi(string(Avalbytes))

	Bvalbytes, err := stub.GetState(B)
	if err != nil {
		return nil, errors.New("Failed to get state")
	}
	if Bvalbytes == nil {
		return nil, errors.New("Entity not found")
	}
	Bval, _ = strconv.Atoi(string(Bvalbytes))

	// Perform the execution
	X, err = strconv.Atoi(args[2])
	Aval = Aval - X
	Bval = Bval + X
	fmt.Printf("Aval = %d, Bval = %d\n", Aval, Bval)

	// Write the state back to the ledger
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return nil, err
	}

	err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// Deletes an entity from state
func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Printf("Running delete")

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 3")
	}

	A := args[0]

	// Delete the key from the state in ledger
	err := stub.DelState(A)
	if err != nil {
		return nil, errors.New("Failed to delete state")
	}

	return nil, nil
}

// Invoke callback representing the invocation of a chaincode
// This chaincode will manage two accounts A and B and will transfer X units from A to B upon invoke
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Printf("Invoke called, determining function")

	// Handle different functions
	if function == "invoke" {
		// Transaction makes payment of X units from A to B
		fmt.Printf("Function is invoke")
		return t.invoke(stub, args)
	} else if function == "init" {
		fmt.Printf("Function is init")
		return t.Init(stub, function, args)
	} else if function == "delete" {
		// Deletes an entity from its state
		fmt.Printf("Function is delete")
		return t.delete(stub, args)
	}

	return nil, errors.New("Received unknown function invocation")
}

func (t *SimpleChaincode) Run(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Printf("Run called, passing through to Invoke (same function)")

	// Handle different functions
	if function == "invoke" {
		// Transaction makes payment of X units from A to B
		fmt.Printf("Function is invoke")
		return t.invoke(stub, args)
	} else if function == "init" {
		fmt.Printf("Function is init")
		return t.Init(stub, function, args)
	} else if function == "delete" {
		// Deletes an entity from its state
		fmt.Printf("Function is delete")
		return t.delete(stub, args)
	}

	return nil, errors.New("Received unknown function invocation")
}

// Query callback representing the query of a chaincode
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Printf("Query called, determining function")

	/*
		********************************************old function body*******************************************
		if function != "query" {
			fmt.Printf("Function is query")
			return nil, errors.New("Invalid query function name. Expecting \"query\"")
		}
			var A string // Entities
			var err error

			if len(args) != 1 {
				return nil, errors.New("Incorrect number of arguments. Expecting name of the person to query")
			}

			A = args[0]

			// Get the state from the ledger
			Avalbytes, err := stub.GetState(A)
			if err != nil {
				jsonResp := "{\"Error\":\"Failed to get state for " + A + "\"}"
				return nil, errors.New(jsonResp)
			}

			if Avalbytes == nil {
				jsonResp := "{\"Error\":\"Nil amount for " + A + "\"}"
				return nil, errors.New(jsonResp)
			}

			jsonResp := "{\"Name\":\"" + A + "\",\"Amount\":\"" + string(Avalbytes) + "\"}"
			fmt.Printf("Query Response:%s\n", jsonResp)
			return Avalbytes, nil
			********************************************old function body*******************************************
	*/

	if function == "query" {
		fmt.Printf("Function is query")
		//return nil, errors.New("Invalid query function name. Expecting \"query\"")
		//}
		var A string // Entities
		var err error

		if len(args) != 1 {
			return nil, errors.New("Incorrect number of arguments. Expecting name of the person to query")
		}

		A = args[0]

		// Get the state from the ledger
		Avalbytes, err := stub.GetState(A)
		if err != nil {
			jsonResp := "{\"Error\":\"Failed to get state for " + A + "\"}"
			return nil, errors.New(jsonResp)
		}

		if Avalbytes == nil {
			jsonResp := "{\"Error\":\"Nil amount for " + A + "\"}"
			return nil, errors.New(jsonResp)
		}

		jsonResp := "{\"Name\":\"" + A + "\",\"Amount\":\"" + string(Avalbytes) + "\"}"
		fmt.Printf("Query Response:%s\n", jsonResp)
		return Avalbytes, nil
	} else if function == "GetCompany" {
		var company Account
		fmt.Println("Getting the company")
		companyBytes, err := stub.GetState(accountPrefix + args[0])
		if err != nil {
			return companyBytes, nil
		}
		
		jsonResp := "{\"Error\":\"Nil amount for " + args[0] + "\"}"
		return nil, errors.New(jsonResp)
		/*company, err := GetCompany(args[0], stub)
		if err != nil {
			fmt.Println("Error from getCompany")
			return nil, err
		} else {
			companyBytes, err1 := json.Marshal(&company)
			if err1 != nil {
				fmt.Println("Error marshalling the company")
				return nil, err1
			}
			fmt.Println("All success, returning the company")
			return companyBytes, nil
		}*/
	}

	return nil, errors.New("Invalid query function name. Expecting \"query\"")

}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

//===========================start============get company info=================================================
func GetCompany(companyID string, stub shim.ChaincodeStubInterface) (Account, error) {
	var company Account
	companyBytes, err := stub.GetState(accountPrefix + companyID)
	if err != nil {
		fmt.Println("Account not found " + companyID)
		return company, errors.New("Account not found " + companyID)
	}

	err = json.Unmarshal(companyBytes, &company)
	if err != nil {
		fmt.Println("Error unmarshalling account " + companyID + "\n err:" + err.Error())
		return company, errors.New("Error unmarshalling account " + companyID)
	}

	return company, nil
}

//===========================end============get company info=================================================

//===========================start============Account creation=================================================
func (t *SimpleChaincode) createAccount(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Creating account")

	// Obtain the username to associate with the account
	if len(args) != 1 {
		fmt.Println("Error obtaining username")
		return nil, errors.New("createAccount accepts a single username argument")
	}
	username := args[0]

	// Build an account object for the user
	var assetIds []string
	suffix := "000A"
	prefix := username + suffix
	var account = Account{ID: username, Prefix: prefix, CashBalance: 10000000.0, AssetsIds: assetIds}
	accountBytes, err := json.Marshal(&account)
	if err != nil {
		fmt.Println("error creating account" + account.ID)
		return nil, errors.New("Error creating account " + account.ID)
	}

	fmt.Println("Attempting to get state of any existing account for " + account.ID)
	existingBytes, err := stub.GetState(accountPrefix + account.ID)
	if err == nil {

		var company Account
		err = json.Unmarshal(existingBytes, &company)
		if err != nil {
			fmt.Println("Error unmarshalling account " + account.ID + "\n--->: " + err.Error())

			if strings.Contains(err.Error(), "unexpected end") {
				fmt.Println("No data means existing account found for " + account.ID + ", initializing account.")
				err = stub.PutState(accountPrefix+account.ID, accountBytes)

				if err == nil {
					fmt.Println("created account" + accountPrefix + account.ID)
					return nil, nil
				} else {
					fmt.Println("failed to create initialize account for " + account.ID)
					return nil, errors.New("failed to initialize an account for " + account.ID + " => " + err.Error())
				}
			} else {
				return nil, errors.New("Error unmarshalling existing account " + account.ID)
			}
		} else {
			fmt.Println("Account already exists for " + account.ID + " " + company.ID)
			return nil, errors.New("Can't reinitialize existing user " + account.ID)
		}
	} else {

		fmt.Println("No existing account found for " + account.ID + ", initializing account.")
		err = stub.PutState(accountPrefix+account.ID, accountBytes)

		if err == nil {
			fmt.Println("created account" + accountPrefix + account.ID)
			return nil, nil
		} else {
			fmt.Println("failed to create initialize account for " + account.ID)
			return nil, errors.New("failed to initialize an account for " + account.ID + " => " + err.Error())
		}

	}

}

func (t *SimpleChaincode) createAccounts(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Creating accounts")

	//  				0
	// "number of accounts to create"
	var err error
	numAccounts, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("error creating accounts with input")
		return nil, errors.New("createAccounts accepts a single integer argument")
	}
	//create a bunch of accounts
	var account Account
	counter := 1
	for counter <= numAccounts {
		var prefix string
		suffix := "000A"
		if counter < 10 {
			prefix = strconv.Itoa(counter) + "0" + suffix
		} else {
			prefix = strconv.Itoa(counter) + suffix
		}
		var assetIds []string
		account = Account{ID: "company" + strconv.Itoa(counter), Prefix: prefix, CashBalance: 10000000.0, AssetsIds: assetIds}
		accountBytes, err := json.Marshal(&account)
		if err != nil {
			fmt.Println("error creating account" + account.ID)
			return nil, errors.New("Error creating account " + account.ID)
		}
		err = stub.PutState(accountPrefix+account.ID, accountBytes)
		counter++
		fmt.Println("created account" + accountPrefix + account.ID)
	}

	fmt.Println("Accounts created")
	return nil, nil

}

//===========================end============Account creation=================================================

//===========================start============standard value =================================================
//lookup tables for last two digits of CUSIP
var seventhDigit = map[int]string{
	1:  "A",
	2:  "B",
	3:  "C",
	4:  "D",
	5:  "E",
	6:  "F",
	7:  "G",
	8:  "H",
	9:  "J",
	10: "K",
	11: "L",
	12: "M",
	13: "N",
	14: "P",
	15: "Q",
	16: "R",
	17: "S",
	18: "T",
	19: "U",
	20: "V",
	21: "W",
	22: "X",
	23: "Y",
	24: "Z",
}

var eigthDigit = map[int]string{
	1:  "1",
	2:  "2",
	3:  "3",
	4:  "4",
	5:  "5",
	6:  "6",
	7:  "7",
	8:  "8",
	9:  "9",
	10: "A",
	11: "B",
	12: "C",
	13: "D",
	14: "E",
	15: "F",
	16: "G",
	17: "H",
	18: "J",
	19: "K",
	20: "L",
	21: "M",
	22: "N",
	23: "P",
	24: "Q",
	25: "R",
	26: "S",
	27: "T",
	28: "U",
	29: "V",
	30: "W",
	31: "X",
}

//===========================end============standard value=================================================

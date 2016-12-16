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
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	//"github.com/test56tester28tt/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type Account struct {
	ID          string  `json:"id"`
	Prefix      string  `json:"prefix"`
	CashBalance float64 `json:"cashBalance"`
}

//============start==========added globle var===============
var cpPrefix = "cp:"
var cpPrefixTest = "cptest:"
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

//===========end======added for account creation ================

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Printf("========================= Init called, initializing chaincode")

	/*if len(args) == 0 {
		return nil, errors.New("Incorrect number of arguments. Expecting 4")
	}*/
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
	fmt.Printf("=========================Aval = %d, Bval = %d\n", Aval, Bval)

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
	fmt.Printf("=========================Running invoke")

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
	fmt.Printf("=========================Aval = %d, Bval = %d\n", Aval, Bval)

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
	fmt.Printf("=========================Running delete")

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
	fmt.Printf("=========================Invoke called, determining function==========val = = " + function)

	// Handle different functions
	if function == "invoke" {
		// Transaction makes payment of X units from A to B
		fmt.Printf("=========================Function is invoke")
		return t.invoke(stub, args)
	} else if function == "init" {
		fmt.Printf("=========================Function is init")
		//return t.Init(stub, function, args)
		return nil, nil
	} else if function == "delete" {
		// Deletes an entity from its state
		fmt.Printf("=========================Function is delete")
		return t.delete(stub, args)
	} else if function == "createAccount" {
		// Deletes an entity from its state
		fmt.Printf("=========================Function is createAccount")
		return t.createAccount(stub, args)
	} else if function == "transaction" {
		fmt.Printf("=========================Function is transaction")
		return t.transaction(stub, args)
	}

	return nil, errors.New("Received unknown function invocation")
}

func (t *SimpleChaincode) Run(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Printf("=========================Run called, passing through to Invoke (same function)")

	// Handle different functions
	if function == "invoke" {
		// Transaction makes payment of X units from A to B
		fmt.Printf("=========================Function is invoke")
		return t.invoke(stub, args)
	} else if function == "init" {
		fmt.Printf("=========================Function is init === calling createAccount function ==")
		//return t.Init(stub, function, args)
		return t.createAccount(stub, args)
	} else if function == "delete" {
		// Deletes an entity from its state
		fmt.Printf("=========================Function is delete")
		return t.delete(stub, args)
	}
	return nil, errors.New("Received unknown function invocation")
}

// Query callback representing the query of a chaincode
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Printf("=========================Query called, determining function")

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
	fmt.Printf("=========================In Query Method=====================val = " + function + "====")
	if function == "query" {
		fmt.Printf("==================Function is query =====================")
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
		fmt.Printf("Query Response =============:%s\n", jsonResp)
		return Avalbytes, nil
	} else if function == "GetCompany" {
		fmt.Println("Getting the company")
		company, err := GetCompany(args[0], stub)
		if err != nil {
			fmt.Println("Error from getCompany")
			return nil, errors.New("User Does not exist")
		} else {
			companyBytes, err1 := json.Marshal(&company)
			if err1 != nil {
				fmt.Println("Error marshalling the company")
				return nil, errors.New("User Does not exist")
			}
			fmt.Println("All success, returning the company")
			return companyBytes, nil
		}
	}
	fmt.Printf("=========================Error in Query=====================")
	return nil, errors.New("Invalid query function name. Expecting \"query\"")
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("=========================Error starting Simple chaincode: %s", err)
	}
}

//===========================start============get company info=================================================
func GetCompany(companyID string, stub shim.ChaincodeStubInterface) (Account, error) {
	var company Account
	companyBytes, err := stub.GetState(accountPrefix + companyID)
	if err != nil {
		fmt.Println("Account not found " + companyID)
		return company, errors.New("Account not found for " + companyID)
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
	if len(args) != 3 {
		fmt.Println("====================Error obtaining username")
		return nil, errors.New("Invalid number of argument")
	}
	username := args[0]
	usertype := args[1]
	suffix := ""
	if usertype == "ADMIN" {
		suffix := "000A"
	} else if usertype == "CORPORATE" {
		suffix := "000C"
	} else if usertype == "NGO" {
		suffix := "000N"
	} else if usertype == "VENDOR" {
		suffix := "000V"
	} else {
		fmt.Println("====================Error obtaining account type")
		return nil, errors.New("Invalid account type")
	}

	amount, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		fmt.Println("===============Invalid Amount" + username)
		return nil, errors.New("Invalid Amount " + args[2] + " for " + username)
	}
	// Build an account object for the user
	prefix := username + suffix
	var account = Account{ID: username, Prefix: prefix, CashBalance: amount}
	accountBytes, err := json.Marshal(&account)
	if err != nil {
		fmt.Println("===============error creating account" + account.ID)
		return nil, errors.New("Error creating account for " + account.ID)
	}

	fmt.Println("==============Attempting to get state of any existing account for " + account.ID + " =Prefix= " + account.Prefix)
	existingBytes, err := stub.GetState(accountPrefix + account.ID)
	if err == nil {

		var useracct Account
		err = json.Unmarshal(existingBytes, &useracct)
		if err != nil {
			fmt.Println("===============Error unmarshalling account " + account.ID + "\n--->: " + err.Error())

			if strings.Contains(err.Error(), "unexpected end") {
				fmt.Println("================No data means existing account found for " + account.ID + ", initializing account.")
				err = stub.PutState(accountPrefix+account.ID, accountBytes)

				if err == nil {
					fmt.Println("================created account" + accountPrefix + account.ID)
					return accountBytes, nil
				} else {
					fmt.Println("==============failed to create initialize account for " + account.ID)
					return nil, errors.New("Failed to initialize an account for " + account.ID + " => " + err.Error())
				}
			} else {
				return nil, errors.New("Error while obtaining existing account " + account.ID)
			}
		} else {
			fmt.Println("=================Account already exists for " + useracct.ID + " " + useracct.ID)
			return nil, errors.New("Account already existing for user " + account.ID)
		}
	} else {

		fmt.Println("==============No existing account found for " + account.ID + ", initializing account.")
		err = stub.PutState(accountPrefix+account.ID, accountBytes)

		if err == nil {
			fmt.Println("============created account" + accountPrefix + account.ID)
			return accountBytes, nil
		} else {
			fmt.Println("==============failed to create initialize account for " + account.ID)
			return nil, errors.New("Failed to initialize an account for " + account.ID + " => " + err.Error())
		}

	}

}

/*func (t *SimpleChaincode) createAccounts(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
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
		account = Account{ID: "company" + strconv.Itoa(counter), Prefix: prefix, CashBalance: 100000.0}
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

}*/

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

//===========================start============transaction function=================================================
func (t *SimpleChaincode) transaction(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("====================Transferring amount to user.=========================")

	if len(args) != 3 {
		return nil, errors.New("Incorrect number of arguments. Expecting commercial paper record")
	}

	var fromCompany Account
	fmt.Println("==============Getting State on fromCompany " + args[0] + "================")
	fromCompanyBytes, err := stub.GetState(accountPrefix + args[0])
	if err != nil {
		fmt.Println("===================Account not found " + args[0] + "================")
		return nil, errors.New("Account not found " + args[0])
	}

	fmt.Println("===============Unmarshalling FromCompany ================")
	err = json.Unmarshal(fromCompanyBytes, &fromCompany)
	if err != nil {
		fmt.Println("===================Error unmarshalling account " + args[0])
		return nil, errors.New("Error unmarshalling account " + args[0])
	}

	var toCompany Account
	fmt.Println("=====================Getting State on ToCompany " + args[1] + "================")
	toCompanyBytes, err := stub.GetState(accountPrefix + args[1])
	if err != nil {
		fmt.Println("Account not found " + args[1] + "================")
		return nil, errors.New("Account not found " + args[1])
	}

	fmt.Println("==================Unmarshalling tocompany================")
	err = json.Unmarshal(toCompanyBytes, &toCompany)
	if err != nil {
		fmt.Println("Error unmarshalling account " + args[1] + "================")
		return nil, errors.New("Error unmarshalling account " + args[1])
	}

	amountToBeTransferred, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		fmt.Println("===================Error converting amount to float " + args[2])
		return nil, errors.New("==============Error converting amount to float " + args[2])
	}

	// If fromCompany doesn't have enough cash to buy the papers
	if fromCompany.CashBalance < amountToBeTransferred {
		fmt.Println("===============The company " + args[1] + "doesn't have enough cash to complete the transaction")
		return nil, errors.New("The company " + args[0] + "doesn't have enough cash to complete the transaction")
	} else {
		fmt.Println("===================The fromCompany has enough money to be transferred amount = " + args[2] + "==========")
	}

	toCompany.CashBalance += amountToBeTransferred
	fromCompany.CashBalance -= amountToBeTransferred

	// Write everything back
	// To Company
	fmt.Println("============= marshalling the toCompany=================")
	toCompanyBytesToWrite, err := json.Marshal(&toCompany)
	if err != nil {
		fmt.Println("=============Error marshalling the toCompany")
		return nil, errors.New("Error marshalling the toCompany")
	}
	fmt.Println("==============Put state on toCompany========amt = %f " + strconv.FormatFloat(toCompany.CashBalance, 'f', 6, 64) + "==========")
	err = stub.PutState(accountPrefix+args[0], toCompanyBytesToWrite)
	if err != nil {
		fmt.Println("===============Error writing the toCompany back")
		return nil, errors.New("Error writing the toCompany back")
	}

	// From company
	fmt.Println("============= marshalling the fromCompany=================")
	fromCompanyBytesToWrite, err := json.Marshal(&fromCompany)
	if err != nil {
		fmt.Println("===============Error marshalling the fromCompany=================")
		return nil, errors.New("Error marshalling the fromCompany")
	}
	fmt.Println("==============Put state on fromCompany amt = %f" + strconv.FormatFloat(fromCompany.CashBalance, 'f', 6, 64) + "==============")
	err = stub.PutState(accountPrefix+args[1], fromCompanyBytesToWrite)
	if err != nil {
		fmt.Println("================Error writing the fromCompany back")
		return nil, errors.New("Error writing the fromCompany back")
	}

	fmt.Println("==================***=== Successfully Transaction completed ====***====================")
	return nil, nil
}

//===========================start============transaction function=================================================

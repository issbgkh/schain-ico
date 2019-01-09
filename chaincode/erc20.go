package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/lib/cid"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// ERC20Chaincode structure
type ERC20Chaincode struct {
}

// Init function
func (e *ERC20Chaincode) Init(stub shim.ChaincodeStubInterface) (res pb.Response) {
	defer func() {
		if r, ok := recover().(error); ok {
			res = shim.Error(r.Error())
		}
	}()

	fcn, args := stub.GetFunctionAndParameters()

	switch fcn {
	case "init":
		res = e.init(stub, args)
	default:
		res = shim.Success(nil)
	}

	return
}

// Invoke function
func (e *ERC20Chaincode) Invoke(stub shim.ChaincodeStubInterface) (res pb.Response) {
	defer func() {
		if r, ok := recover().(error); ok {
			res = shim.Error(r.Error())
		}
	}()

	fcn, args := stub.GetFunctionAndParameters()

	switch fcn {
	case "totalSupply":
		res = e.totalSupply(stub, args)
	case "balanceOf":
		res = e.balanceOf(stub, args)
	case "allowance":
		res = e.allowance(stub, args)
	case "transfer":
		res = e.transfer(stub, args)
	case "approve":
		res = e.approve(stub, args)
	case "transferFrom":
		res = e.transferFrom(stub, args)
	default:
		res = shim.Error("invalid function name")
	}

	return
}

const totalSupply = 100000000

type transfer struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Value uint64 `json:"value"`
}

type approval struct {
	Owner   string `json:"owner"`
	Spender string `json:"spender"`
	Value   uint64 `json:"value"`
}

func (e *ERC20Chaincode) init(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("invalid arguments")
	}

	owner := args[0]

	setBalance(stub, owner, totalSupply)
	emitEvent(stub, "Transfer", getTransfer("", owner, totalSupply))

	return shim.Success(nil)
}

func (e *ERC20Chaincode) totalSupply(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 0 {
		return shim.Error("invalid arguments")
	}

	return shim.Success(u2b(totalSupply))
}

func (e *ERC20Chaincode) balanceOf(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("invalid arguments")
	}

	owner := args[0]

	balance := getBalance(stub, owner)

	return shim.Success(u2b(balance))
}

func (e *ERC20Chaincode) allowance(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("invalid arguments")
	}

	owner := args[0]
	spender := args[1]

	allowance := getAllowance(stub, owner, spender)

	return shim.Success(u2b(allowance))
}

func (e *ERC20Chaincode) transfer(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("invalid arguments")
	}

	sender := getSender(stub)
	to := args[0]
	value := s2u(args[1])

	setBalance(stub, sender, sub(getBalance(stub, sender), value))
	setBalance(stub, to, add(getBalance(stub, to), value))
	emitEvent(stub, "Transfer", getTransfer(sender, to, value))

	return shim.Success([]byte("true"))
}

func (e *ERC20Chaincode) approve(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("invalid arguments")
	}

	sender := getSender(stub)
	spender := args[0]
	value := s2u(args[1])

	setAllowance(stub, sender, spender, value)
	emitEvent(stub, "Approval", getApproval(sender, spender, value))

	return shim.Success([]byte("true"))
}

func (e *ERC20Chaincode) transferFrom(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 3 {
		return shim.Error("invalid arguments")
	}

	sender := getSender(stub)
	from := args[0]
	to := args[1]
	value := s2u(args[2])

	setBalance(stub, from, sub(getBalance(stub, from), value))
	setAllowance(stub, from, sender, sub(getAllowance(stub, from, sender), value))
	setBalance(stub, to, add(getBalance(stub, to), value))
	emitEvent(stub, "Transfer", getTransfer(from, to, value))

	return shim.Success([]byte("true"))
}

func getSender(stub shim.ChaincodeStubInterface) string {
	cert, err := cid.GetX509Certificate(stub)

	checkError(err)

	return cert.Subject.CommonName
}

func getBalance(stub shim.ChaincodeStubInterface, owner string) uint64 {
	// TODO: json format could be better for CouchDB?
	key := fmt.Sprintf("balance::%s", owner)

	balance, err := stub.GetState(key)

	checkError(err)

	if balance == nil {
		return 0
	}

	return b2u(balance)
}

func setBalance(stub shim.ChaincodeStubInterface, owner string, balance uint64) {
	key := fmt.Sprintf("balance::%s", owner)

	err := stub.PutState(key, u2b(balance))

	checkError(err)
}

func getAllowance(stub shim.ChaincodeStubInterface, owner string, spender string) uint64 {
	key := fmt.Sprintf("allowance::%s::%s", owner, spender)

	allowance, err := stub.GetState(key)

	checkError(err)

	if allowance == nil {
		return 0
	}

	return b2u(allowance)
}

func setAllowance(stub shim.ChaincodeStubInterface, owner string, spender string, allowance uint64) {
	key := fmt.Sprintf("allowance::%s::%s", owner, spender)

	err := stub.PutState(key, u2b(allowance))

	checkError(err)
}

func getTransfer(from string, to string, value uint64) []byte {
	transfer, err := json.Marshal(&transfer{from, to, value})

	checkError(err)

	return transfer
}

func getApproval(owner string, spender string, value uint64) []byte {
	approval, err := json.Marshal(&approval{owner, spender, value})

	checkError(err)

	return approval
}

func emitEvent(stub shim.ChaincodeStubInterface, name string, payload []byte) {
	err := stub.SetEvent(name, payload)

	checkError(err)
}

func main() {
	err := shim.Start(new(ERC20Chaincode))

	if err != nil {
		fmt.Printf("Error: %s", err)
	}
}

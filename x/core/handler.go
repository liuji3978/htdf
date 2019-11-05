package htdfservice

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	vmcore "github.com/orientwalt/htdf/evm/core"
	"github.com/orientwalt/htdf/evm/state"
	"github.com/orientwalt/htdf/evm/vm"
	appParams "github.com/orientwalt/htdf/params"
	sdk "github.com/orientwalt/htdf/types"
	"github.com/orientwalt/htdf/x/auth"
)

//
type SendTxResp struct {
	ErrCode         sdk.CodeType `json:"code"`
	ErrMsg          string       `json:"message"`
	ContractAddress string       `json:"contract_address"`
	EvmOutput       string       `json:"evm_output"`
}

//
func (rsp SendTxResp) String() string {
	rsp.ErrMsg = sdk.GetErrMsg(rsp.ErrCode)
	data, _ := json.Marshal(&rsp)
	return string(data)
}

// New HTDF Message Handler
// connected to handler.go
// HandleMsgSendFrom, HandleMsgAdd upgraded to EVM version
// commented by junying, 2019-08-21
func NewHandler(accountKeeper auth.AccountKeeper,
	feeCollectionKeeper auth.FeeCollectionKeeper,
	keyStorage *sdk.KVStoreKey,
	keyCode *sdk.KVStoreKey) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {

		switch msg := msg.(type) {
		case MsgSendFrom:
			return HandleMsgSendFrom(ctx, accountKeeper, feeCollectionKeeper, keyStorage, keyCode, msg)
		default:
			return HandleUnknownMsg(msg)
		}
	}

}

// junying-todo, 2019-08-26
func HandleUnknownMsg(msg sdk.Msg) sdk.Result {
	var sendTxResp SendTxResp
	fmt.Printf("msgType error|mstType=%v\n", msg.Type())
	sendTxResp.ErrCode = sdk.ErrCode_Param
	return sdk.Result{Code: sendTxResp.ErrCode, Log: sendTxResp.String()}
}

// junying-todo, 2019-08-26
func HandleMsgSendFrom(ctx sdk.Context,
	accountKeeper auth.AccountKeeper,
	feeCollectionKeeper auth.FeeCollectionKeeper,
	keyStorage *sdk.KVStoreKey,
	keyCode *sdk.KVStoreKey,
	msg MsgSendFrom) sdk.Result {
	var sendTxResp SendTxResp
	fmt.Printf("FeeTotal1=%v\n", feeCollectionKeeper.GetCollectedFees(ctx))

	if !msg.To.Empty() {
		//open smart contract

		fmt.Printf("openContract\n")
		fmt.Println("HandleMsgSendFrom0:ctx.GasMeter().GasConsumed()", ctx.GasMeter().GasConsumed())
		evmOutput, gasUsed, genErr := HandleOpenContract(ctx, accountKeeper, feeCollectionKeeper, keyStorage, keyCode, msg)
		fmt.Println("HandleMsgSendFrom1:ctx.GasMeter().GasConsumed()", ctx.GasMeter().GasConsumed())
		if genErr != nil {
			fmt.Printf("openContract error|err=%s\n", genErr)
			sendTxResp.ErrCode = sdk.ErrCode_OpenContract
			return sdk.Result{Code: sendTxResp.ErrCode, Log: sendTxResp.String()}
		}

		sendTxResp.EvmOutput = evmOutput
		return sdk.Result{Code: sendTxResp.ErrCode, Log: sendTxResp.String(), GasUsed: gasUsed}

	} else {
		// new smart contract
		fmt.Printf("create contract\n")
		contractAddr, gasUsed, err := HandleCreateContract(ctx, accountKeeper, feeCollectionKeeper, keyStorage, keyCode, msg)
		if err != nil {
			fmt.Printf("CreateContract error|err=%s\n", err)
			sendTxResp.ErrCode = sdk.ErrCode_CreateContract
			return sdk.Result{Code: sendTxResp.ErrCode, Log: sendTxResp.String()}
		}

		sendTxResp.ContractAddress = contractAddr
		return sdk.Result{Code: sendTxResp.ErrCode, Log: sendTxResp.String(), GasUsed: gasUsed}
	}

	// fmt.Printf("FeeTotal2=%v\n", feeCollectionKeeper.GetCollectedFees(ctx))
}

// junying-todo, 2019-08-26
//
func HandleOpenContract(ctx sdk.Context,
	accountKeeper auth.AccountKeeper,
	feeCollectionKeeper auth.FeeCollectionKeeper,
	keyStorage *sdk.KVStoreKey,
	keyCode *sdk.KVStoreKey,
	msg MsgSendFrom) (evmOutput string, gasUsed uint64, err error) {

	fmt.Printf("Handling MsgSendFrom with No Contract.\n")
	fmt.Println(" HandleOpenContract0:ctx.GasMeter().GasConsumed()", ctx.GasMeter().GasConsumed())
	stateDB, err := state.NewCommitStateDB(ctx, &accountKeeper, keyStorage, keyCode)
	if err != nil {
		fmt.Printf("newStateDB error\n")
		return
	}

	fromAddress := sdk.ToEthAddress(msg.From)
	toAddress := sdk.ToEthAddress(msg.To)

	fmt.Printf("fromAddr|appFormat=%s|ethFormat=%s|\n", msg.From.String(), fromAddress.String())
	fmt.Printf("toAddress|appFormat=%s|ethFormat=%s|\n", msg.To.String(), toAddress.String())

	fmt.Printf("fromAddress|testBalance=%v\n", stateDB.GetBalance(fromAddress))
	fmt.Printf("fromAddress|nonce=%d\n", stateDB.GetNonce(fromAddress))

	config := appParams.MainnetChainConfig
	logConfig := vm.LogConfig{}
	structLogger := vm.NewStructLogger(&logConfig)
	vmConfig := vm.Config{Debug: true, Tracer: structLogger /*, JumpTable: vm.NewByzantiumInstructionSet()*/}

	evmCtx := vmcore.NewEVMContext(msg, &fromAddress, uint64(ctx.BlockHeight()))
	evm := vm.NewEVM(evmCtx, stateDB, config, vmConfig)
	contractRef := vm.AccountRef(fromAddress)

	inputCode, err := hex.DecodeString(msg.Data)
	if err != nil {
		fmt.Printf("DecodeString error\n")
		return
	}

	fmt.Printf("inputCode=%s\n", hex.EncodeToString(inputCode))

	transferAmount := msg.Amount.AmountOf(sdk.DefaultDenom).BigInt()

	fmt.Printf("transferAmount: %d\n", transferAmount)
	st := NewStateTransition(evm, msg, stateDB)

	fmt.Printf("gas=%d|gasPrice=%d|gasLimit=%d\n", msg.Gas, msg.GasPrice, msg.GasLimit)

	// commented by junying, 2019-08-22
	// subtract gaslimit*gasprice from sender
	err = st.buyGas()
	if err != nil {
		fmt.Printf("buyGas error|err=%s\n", err)
		return
	}

	// Intrinsic gas calc
	// commented by junying, 2019-08-22
	// default non-contract tx gas: 21000
	// default contract tx gas: 53000 + f(tx.data)
	itrsGas, err := IntrinsicGas(inputCode, true)
	fmt.Printf("itrsGas|gas=%d\n", itrsGas)
	// commented by junying, 2019-08-22
	// check if tx.gas >= calculated gas
	err = st.useGas(itrsGas)
	if err != nil {
		fmt.Printf("useGas error|err=%s\n", err)
		return
	}

	// commented by junying, 2019-08-22
	// 1. cantransfer check
	// 2. create receiver account if no exists
	// 3. execute contract & calculate gas
	fmt.Println(" HandleOpenContract1:ctx.GasMeter().GasConsumed()", ctx.GasMeter().GasConsumed())
	outputs, gasLeftover, vmerr := evm.Call(contractRef, toAddress, inputCode, st.gas, transferAmount)
	fmt.Println(" HandleOpenContract2:ctx.GasMeter().GasConsumed()", ctx.GasMeter().GasConsumed())
	if err != nil {
		fmt.Printf("evm call error|err=%s\n", vmerr)
		return
	}

	st.gas = gasLeftover
	// junying-todo, 2019-08-22
	// refund(add) remaining to sender
	st.refundGas()

	fmt.Printf("gasUsed=%d\n", st.gasUsed())

	// gasUsedValue
	gasUsedValue := new(big.Int).Mul(new(big.Int).SetUint64(st.gasUsed()), st.gasPrice)
	fmt.Printf("gasUsedValue=%s\n", gasUsedValue.String())

	// junying-todo, 2019-08-22
	// this function is used to collect all kinds of budget including fee + blk rewards for the next block reward
	feeCollectionKeeper.AddCollectedFees(ctx, sdk.Coins{sdk.NewCoin(sdk.DefaultDenom, sdk.NewIntFromBigInt(gasUsedValue))})
	fmt.Println(" HandleOpenContract3:ctx.GasMeter().GasConsumed()", ctx.GasMeter().GasConsumed())
	fmt.Printf("evm call end|outputs=%x\n", outputs)

	stateDB.Commit(false)
	fmt.Println(" HandleOpenContract4:ctx.GasMeter().GasConsumed()", ctx.GasMeter().GasConsumed())
	return hex.EncodeToString(outputs), st.gasUsed(), nil
}

// junying-todo, 2019-08-26
func HandleCreateContract(ctx sdk.Context,
	accountKeeper auth.AccountKeeper,
	feeCollectionKeeper auth.FeeCollectionKeeper,
	keyStorage *sdk.KVStoreKey,
	keyCode *sdk.KVStoreKey,
	msg MsgSendFrom) (evmOutput string, gasUsed uint64, err error) {

	stateDB, err := state.NewCommitStateDB(ctx, &accountKeeper, keyStorage, keyCode)
	if err != nil {
		fmt.Printf("newStateDB error\n")
		return
	}

	fromAddress := sdk.ToEthAddress(msg.From)
	toAddress := sdk.ToEthAddress(msg.To)

	fmt.Printf("fromAddr|appFormat=%s|ethFormat=%s|\n", msg.From.String(), fromAddress.String())
	fmt.Printf("toAddress|appFormat=%s|ethFormat=%s|\n", msg.To.String(), toAddress.String())
	fmt.Printf("fromAddress|Balance=%v\n", stateDB.GetBalance(fromAddress))

	config := appParams.MainnetChainConfig
	logConfig := vm.LogConfig{}
	structLogger := vm.NewStructLogger(&logConfig)
	vmConfig := vm.Config{Debug: true, Tracer: structLogger /*, JumpTable: vm.NewByzantiumInstructionSet()*/}

	fmt.Printf("fromAddress|nonce=%d\n", stateDB.GetNonce(fromAddress))

	evmCtx := vmcore.NewEVMContext(msg, &fromAddress, uint64(ctx.BlockHeight()))
	evm := vm.NewEVM(evmCtx, stateDB, config, vmConfig)
	contractRef := vm.AccountRef(fromAddress)

	fmt.Printf("blockHeight=%d|IsHomestead=%v|IsDAOFork=%v|IsEIP150=%v|IsEIP155=%v|IsEIP158=%v|IsByzantium=%v\n", ctx.BlockHeight(), evm.ChainConfig().IsHomestead(evm.BlockNumber),
		evm.ChainConfig().IsDAOFork(evm.BlockNumber), evm.ChainConfig().IsEIP150(evm.BlockNumber),
		evm.ChainConfig().IsEIP155(evm.BlockNumber), evm.ChainConfig().IsEIP158(evm.BlockNumber),
		evm.ChainConfig().IsByzantium(evm.BlockNumber))

	inputCode, err := hex.DecodeString(msg.Data)
	if err != nil {
		fmt.Printf("DecodeString error\n")
		return
	}

	fmt.Printf("inputCode=%s\n", hex.EncodeToString(inputCode))

	st := NewStateTransition(evm, msg, stateDB)

	fmt.Printf("gas=%d|gasPrice=%d|gasLimit=%d\n", msg.Gas, msg.GasPrice, msg.GasLimit)

	err = st.buyGas()
	if err != nil {
		fmt.Printf("buyGas error|err=%s\n", err)
		return
	}

	//Intrinsic gas calc
	itrsGas, err := IntrinsicGas(inputCode, true)
	fmt.Printf("itrsGas|gas=%d\n", itrsGas)
	err = st.useGas(itrsGas)
	if err != nil {
		fmt.Printf("useGas error|err=%s\n", err)
		return
	}

	_, contractAddr, gasLeftover, vmerr := evm.Create(contractRef, inputCode, st.gas, big.NewInt(0))
	if vmerr != nil {
		fmt.Printf("evm Create error|err=%s\n", vmerr)
		return
	}
	st.gas = gasLeftover

	st.refundGas()

	fmt.Printf("gasUsed=%d\n", st.gasUsed())

	// gasUsedValue
	gasUsedValue := new(big.Int).Mul(new(big.Int).SetUint64(st.gasUsed()), st.gasPrice)
	fmt.Printf("gasUsedValue=%s\n", gasUsedValue.String())
	feeCollectionKeeper.AddCollectedFees(ctx, sdk.Coins{sdk.NewCoin(sdk.DefaultDenom, sdk.NewIntFromBigInt(gasUsedValue))})

	fmt.Printf("Create contract ok,contractAddr|appFormat=%s|ethFormat=%s\n", sdk.ToAppAddress(contractAddr).String(), contractAddr.String())

	stateDB.Commit(false)

	return sdk.ToAppAddress(contractAddr).String(), st.gasUsed(), nil
}

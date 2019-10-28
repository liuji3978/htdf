package rest

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/orientwalt/htdf/client/context"
	"github.com/orientwalt/htdf/client/rpc"
	"github.com/orientwalt/htdf/codec"
	sdk "github.com/orientwalt/htdf/types"
	sdkRest "github.com/orientwalt/htdf/types/rest"
	"github.com/orientwalt/htdf/accounts"
	"github.com/orientwalt/htdf/accounts/keystore"
	"github.com/orientwalt/htdf/utils/unit_convert"
	"github.com/spf13/viper"
	"github.com/tendermint/tmlibs/cli"
	"net/http"
	"path/filepath"

	"github.com/orientwalt/htdf/x/auth"
	"github.com/orientwalt/htdf/x/core"
)

func AccountListRequestHandlerFn(w http.ResponseWriter, r *http.Request) {

	var am *accounts.Manager

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&am)

	// Assemble the account manager and supported backends
	rootDir := viper.GetString(cli.HomeFlag)
	defaultKeyStoreHome := filepath.Join(rootDir, "keystores")
	backends := []accounts.Backend{
		keystore.NewKeyStore(defaultKeyStoreHome),
	}

	am = accounts.NewManager(backends...)
	if am == nil {

		sdkRest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	bJsonFormat := false
	vars := r.URL.Query()
	_, ok := vars["jsonformat"]
	if ok {
		bJsonFormat = true

		//fmt.Printf("jsonformat=%s\n", jsonformat)
	}

	var index int
	if bJsonFormat == false {
		for _, wallet := range am.Wallets() {
			for _, account := range wallet.Accounts() {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(fmt.Sprintf("Account #%d: {%s}\n", index, account.Address)))
				index++
			}
		}
	} else {
		var accounts []accounts.Account
		for _, wallet := range am.Wallets() {
			for _, account := range wallet.Accounts() {
				accounts = append(accounts, account)
			}

		}
		data, err := json.Marshal(accounts)
		if err != nil {
			fmt.Printf("Marshal error|err=%s\n", err)
			sdkRest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.Write(data)
	}
}

type AccountTxsReq struct {
	Address    string `json:"address"`
	FromHeight int64  `json:"fromHeight"`
	EndHeight  int64  `json:"endHeight"`
	Flag       int64  `json:"flag"`
}

func parseTx(cdc *codec.Codec, txBytes []byte) (sdk.Tx, error) {
	var tx auth.StdTx

	err := cdc.UnmarshalBinaryLengthPrefixed(txBytes, &tx)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

//GetAccountTxsFn
//@param Address:  address of account
//@param FromHeight: query tx in the block height range [fromHeight,endHeight];  when FromHeight and endHeight is 0,  range is [CHAIN_HEIGHT-800, CHAIN_HEIGHT]
//@param EndHeight:  query tx in the block height range [fromHeight,endHeight]   when FromHeight and endHeight is 0,  range is [CHAIN_HEIGHT-800, CHAIN_HEIGHT]
//@param Flag:   query flag; 0, address appears both in fromAddress and toAddress; 1, address appears in fromAddress;  2, address appears in toAddress
func GetAccountTxsFn(cliCtx context.CLIContext, cdc *codec.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var accountTxsReq AccountTxsReq
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&accountTxsReq)
		if err != nil {
			fmt.Printf("Decode error|err=%s\n", err)
			sdkRest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		//fmt.Printf("accountTxsReq=%v\n", accountTxsReq)

		//check address
		_, err = sdk.AccAddressFromBech32(accountTxsReq.Address)
		if err != nil {
			fmt.Printf("AccAddressFromBech32 error|err=%s\n", err)
			sdkRest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		fromHeight := accountTxsReq.FromHeight
		endHeight := accountTxsReq.EndHeight

		chainHeight, err := rpc.GetChainHeight(cliCtx)
		if err != nil {
			fmt.Printf("GetChainHeight error|err=%s\n", err)
			sdkRest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		//fmt.Printf("chainHeight=%d\n", chainHeight)

		//from the chain height
		scope := int64(1000)
		if fromHeight == 0 && endHeight == 0 {
			endHeight = chainHeight
			fromHeight = chainHeight - scope
		}

		//correct height parameter
		if fromHeight <= 0 {
			fromHeight = 1
		}

		if endHeight > chainHeight || endHeight == 0 {
			endHeight = chainHeight
		}

		//fmt.Printf("final|fromHeight=%d|endHeight=%d\n", fromHeight, endHeight)

		// get the node
		node, err := cliCtx.GetNode()
		if err != nil {
			fmt.Printf("getNode error|err=%s\n", err)
			sdkRest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		var result ResultAccountTxs
		result.ChainHeight = chainHeight
		result.FromHeight = fromHeight
		result.EndHeight = endHeight

		for height := fromHeight; height <= endHeight; height++ {
			//get Block info
			resultBlock, err := node.Block(&height)
			if err != nil {
				fmt.Printf("get block error|err=%s\n", err)
				sdkRest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}

			//fmt.Printf("currHeight=%d\n", height)
			//fmt.Printf("txTotal=%d\n", len(resultBlock.Block.Txs))

			for _, tx := range resultBlock.Block.Txs {
				sdkTx, err := parseTx(cdc, tx)
				if err != nil {
					fmt.Printf("parseTx error|err=%s\n", err)
					sdkRest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
					return
				}

				switch iMsg := sdkTx.(type) {
				case auth.StdTx:

					for _, msg := range iMsg.GetMsgs() {
						//fmt.Printf("msg|route=%s|type=%s\n", msg.Route(), msg.Type())

						switch msg := msg.(type) {
						case htdfservice.MsgSendFrom:
							//fmt.Printf("msg|from=%s|to=%s\n", msg.From, msg.To)

							if (accountTxsReq.Flag == 0 && (msg.From.String() == accountTxsReq.Address || msg.To.String() == accountTxsReq.Address)) ||
								(accountTxsReq.Flag == 1 && msg.From.String() == accountTxsReq.Address) ||
								(accountTxsReq.Flag == 2 && msg.To.String() == accountTxsReq.Address) {
								//fmt.Printf("msg found|from=%s|to=%s\n", msg.From, msg.To)
								var displayTx DisplayTx
								displayTx.From = msg.From
								displayTx.To = msg.To
								displayTx.Amount = unit_convert.DefaultCoinsToBigCoins(msg.Amount)
								displayTx.Hash = hex.EncodeToString(tx.Hash())
								displayTx.Height = height
								displayTx.Time = resultBlock.BlockMeta.Header.Time.Local().Format("2006-01-02 15:04:05")
								displayTx.Memo = iMsg.Memo
								result.ArrTx = append(result.ArrTx, displayTx)
							}

						default:
							fmt.Printf("ignore type|type=%s|route=%s\n", msg.Type(), msg.Route())
							continue
						}
					}

				default:
					fmt.Printf("unknown type: %+v\n", iMsg)
				}
			}
		}

		sdkRest.PostProcessResponse(w, cdc, &result, cliCtx.Indent)

	}
}
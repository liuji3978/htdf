package rest

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/orientwalt/htdf/client"
	"github.com/orientwalt/htdf/client/tx"
	sdk "github.com/orientwalt/htdf/types"
	"github.com/orientwalt/htdf/types/rest"
	"github.com/orientwalt/htdf/x/bank/types"
)

// SendReq defines the properties of a send request's body.
type SendReq struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	Amount  sdk.Coins    `json:"amount" yaml:"amount"`
}

// NewSendRequestHandlerFn returns an HTTP REST handler for creating a MsgSend
// transaction.
func NewSendRequestHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		bech32Addr := vars["address"]

		toAddr, err := sdk.AccAddressFromBech32(bech32Addr)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		var req SendReq
		if !rest.ReadRESTReq(w, r, clientCtx.LegacyAmino, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		msg := types.NewMsgSend(fromAddr, toAddr, req.Amount)
		tx.WriteGeneratedTxResponse(clientCtx, w, req.BaseReq, msg)
	}
}

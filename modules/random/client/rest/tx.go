package rest

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/orientwalt/htdf/client"
	"github.com/orientwalt/htdf/client/tx"
	"github.com/orientwalt/htdf/types/rest"

	"github.com/orientwalt/htdf/modules/random/types"
)

func registerTxRoutes(cliCtx client.Context, r *mux.Router) {
	// request rands
	r.HandleFunc("/random/randoms", requestRandomHandlerFn(cliCtx)).Methods("POST")
}

// HTTP request handler to request random
func requestRandomHandlerFn(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RequestRandomReq
		if !rest.ReadRESTReq(w, r, cliCtx.LegacyAmino, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		// create the MsgRequestRandom message
		msg := types.NewMsgRequestRandom(req.Consumer, req.BlockInterval, req.Oracle, req.ServiceFeeCap)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

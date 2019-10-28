package rest

import (
	"encoding/json"
	"fmt"

	"github.com/orientwalt/htdf/types/rest"
	"net/http"
	"path/filepath"

	"github.com/orientwalt/htdf/accounts/keystore"
	"github.com/spf13/viper"
	"github.com/tendermint/tmlibs/cli"
)

type newaccountBody struct {
	Password string `json:"password"`
}

func NewAccountRequestHandlerFn(w http.ResponseWriter, r *http.Request) {
	var req newaccountBody

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	rootDir := viper.GetString(cli.HomeFlag)
	defaultKeyStoreHome := filepath.Join(rootDir, "keystores")
	address, _, err := keystore.StoreKey(defaultKeyStoreHome, req.Password)

	if err != nil {
		rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf("{\"address\": \"%s\"}", address)))

	return
}
package simulation

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/orientwalt/htdf/crypto/keys/secp256k1"
	"github.com/orientwalt/htdf/simapp"
	sdk "github.com/orientwalt/htdf/types"
	"github.com/orientwalt/htdf/types/kv"

	"github.com/orientwalt/htdf/modules/record/types"
)

var (
	creatorPk1   = secp256k1.GenPrivKey().PubKey()
	creatorAddr1 = sdk.AccAddress(creatorPk1.Address())
)

func TestDecodeStore(t *testing.T) {
	cdc, _ := simapp.MakeCodecs()
	dec := NewDecodeStore(cdc)

	txHash := make([]byte, 32)
	_, _ = rand.Read(txHash)
	record := types.NewRecord(txHash, nil, creatorAddr1)

	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: types.GetRecordKey(txHash), Value: cdc.MustMarshalBinaryBare(&record)},
			{Key: []byte{0x99}, Value: []byte{0x99}},
		},
	}
	tests := []struct {
		name        string
		expectedLog string
	}{
		{"Record", fmt.Sprintf("%v\n%v", record, record)},
		{"other", ""},
	}

	for i, tt := range tests {
		i, tt := i, tt
		t.Run(tt.name, func(t *testing.T) {
			switch i {
			case len(tests) - 1:
				require.Panics(t, func() { dec(kvPairs.Pairs[i], kvPairs.Pairs[i]) }, tt.name)
			default:
				require.Equal(t, tt.expectedLog, dec(kvPairs.Pairs[i], kvPairs.Pairs[i]), tt.name)
			}
		})
	}
}

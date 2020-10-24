package types

import (
	sdkerrors "github.com/orientwalt/htdf/types/errors"
)

// mint module sentinel errors
var (
	ErrInvalidMintInflation = sdkerrors.Register(ModuleName, 2, "invalid mint inflation")
	ErrInvalidMintDenom     = sdkerrors.Register(ModuleName, 3, "invalid mint denom")
)

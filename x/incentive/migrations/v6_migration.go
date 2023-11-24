package migrations

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/elys-network/elys/x/incentive/types"
)

func (m Migrator) V6Migration(ctx sdk.Context) error {
	m.keeper.SetParams(ctx, types.NewParams())
	return nil
}
package upgrade

import (
	"fmt"
	"strconv"

	sdk "github.com/orientwalt/htdf/types"
)

// do switch
func EndBlocker(ctx sdk.Context, uk Keeper) (tags sdk.Tags) {

	ctx = ctx.WithLogger(ctx.Logger().With("handler", "endBlock").With("module", "iris/upgrade"))

	tags = sdk.NewTags()
	upgradeConfig, ok := uk.protocolKeeper.GetUpgradeConfig(ctx)
	if ok {

		versionIDstr := strconv.FormatUint(upgradeConfig.Protocol.Version, 10)
		uk.metrics.Upgrade.Set(float64(upgradeConfig.Protocol.Version))

		validator, found := uk.sk.GetValidatorByConsAddr(ctx, (sdk.ConsAddress)(ctx.BlockHeader().ProposerAddress))
		if !found {
			panic(fmt.Sprintf("validator with consensus-address %s not found", (sdk.ConsAddress)(ctx.BlockHeader().ProposerAddress).String()))
		}
		fmt.Print("---------blocker header version.app-------", ctx.BlockHeader().Version.App, "\n")
		fmt.Print("---------upgradeConfig.Protocol.Version-------", upgradeConfig.Protocol.Version, "\n")
		if ctx.BlockHeader().Version.App == upgradeConfig.Protocol.Version {
			uk.SetSignal(ctx, upgradeConfig.Protocol.Version, validator.ConsAddress().String())
			uk.metrics.Signal.With(ValidatorLabel, validator.ConsAddress().String(), VersionLabel, versionIDstr).Set(1)

			ctx.Logger().Info("Validator has downloaded the latest software ",
				"validator", validator.GetOperator().String(), "version", upgradeConfig.Protocol.Version)

		} else {

			ok := uk.DeleteSignal(ctx, upgradeConfig.Protocol.Version, validator.ConsAddress().String())
			uk.metrics.Signal.With(ValidatorLabel, validator.ConsAddress().String(), VersionLabel, versionIDstr).Set(0)

			if ok {
				ctx.Logger().Info("Validator has restarted the old software ",
					"validator", validator.GetOperator().String(), "version", upgradeConfig.Protocol.Version)
			}
		}
		fmt.Print("-----------------upgrade height-----------------	", upgradeConfig.Protocol.Height, "\n")
		fmt.Print("-----------------upgrade version-----------------	", upgradeConfig.Protocol.Version, "\n")
		fmt.Print("-----------------upgradeConfig.Protocol.Threshold-----------------	", upgradeConfig.Protocol.Threshold, "\n")
		if uint64(ctx.BlockHeight())+1 == upgradeConfig.Protocol.Height {
			success := tally(ctx, upgradeConfig.Protocol.Version, uk, upgradeConfig.Protocol.Threshold)
			fmt.Print("-----------------upgrade status-----------------	", success, "\n")
			fmt.Print("-----------------upgrade version-----------------	", upgradeConfig.Protocol.Version, "\n")
			if success {
				ctx.Logger().Info("Software Upgrade is successful.", "version", upgradeConfig.Protocol.Version)
				uk.protocolKeeper.SetCurrentVersion(ctx, upgradeConfig.Protocol.Version)
			} else {
				ctx.Logger().Info("Software Upgrade is failure.", "version", upgradeConfig.Protocol.Version)
				uk.protocolKeeper.SetLastFailedVersion(ctx, upgradeConfig.Protocol.Version)
			}

			uk.AddNewVersionInfo(ctx, NewVersionInfo(upgradeConfig, success))
			uk.protocolKeeper.ClearUpgradeConfig(ctx)
		}
	} else {
		uk.metrics.Upgrade.Set(float64(0))
	}

	tags = tags.AppendTag(sdk.AppVersionTag, (strconv.FormatUint(uk.protocolKeeper.GetCurrentVersion(ctx), 10)))

	return tags
}
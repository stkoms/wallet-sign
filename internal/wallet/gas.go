package wallet

import (
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"golang.org/x/xerrors"

	"wallet-sign/internal/chain/types"
	"wallet-sign/internal/vapi"
)

func CapGasFee(msg *types.Message) {
	log.Debugf("CapGasFee: capping gas fee for message to %s", msg.To)

	var maxFee abi.TokenAmount
	if maxFee.Int == nil || maxFee.Equals(big.Zero()) {
		maxFee, _ = big.FromString("10000000000000000000")
	}

	gl := types.NewInt(uint64(msg.GasLimit))
	totalFee := types.BigMul(msg.GasFeeCap, gl)

	if totalFee.LessThanEqual(maxFee) {
		log.Debugf("CapGasFee: total fee %s <= max fee %s, no capping needed", totalFee, maxFee)
		return
	}

	msg.GasFeeCap = big.Div(maxFee, gl)
	msg.GasPremium = big.Min(msg.GasFeeCap, msg.GasPremium) // cap premium at FeeCap

	log.Infof("CapGasFee: capped gas fee from %s to %s", totalFee, msg.GasFeeCap)
}

func SetGas(api *vapi.Node, msg *types.Message) error {
	log.Infof("SetGas: estimating gas for message from %s to %s", msg.From, msg.To)

	gasLimit, err := api.GasEstimateGasLimit(msg)
	if err != nil {
		log.Errorf("SetGas: failed to estimate gas limit: %v", err)
		return err
	}
	// Apply 1.25x overestimation to gas limit for safety margin
	msg.GasLimit = gasLimit.GasLimit
	log.Debugf("SetGas: gas limit set to %d", msg.GasLimit)

	gasPremium, err := api.GasEstimateGasPremium(10, msg.From, msg.GasLimit)
	if err != nil {
		log.Errorf("SetGas: failed to estimate gas premium: %v", err)
		return xerrors.Errorf("estimating gas price: %w", err)
	}
	msg.GasPremium = gasPremium
	log.Debugf("SetGas: gas premium set to %s", gasPremium)

	if msg.GasFeeCap == types.EmptyInt || types.BigCmp(msg.GasFeeCap, types.NewInt(0)) == 0 {
		feeCap, err := api.GasEstimateFeeCap(msg, 20)
		if err != nil {
			log.Errorf("SetGas: failed to estimate fee cap: %v", err)
			return nil
		}
		msg.GasFeeCap = feeCap
		log.Debugf("SetGas: gas fee cap set to %s", feeCap)
	}

	CapGasFee(msg)
	log.Infof("SetGas: successfully set gas parameters: limit=%d premium=%s feecap=%s",
		msg.GasLimit, msg.GasPremium, msg.GasFeeCap)
	return nil
}

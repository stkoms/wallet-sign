package wallet

import (
	"fmt"

	"github.com/filecoin-project/go-state-types/crypto"

	"wallet-sign/internal/chain/types"
)

// sigTypeForKeyType 将密钥类型转换为签名类型
// 支持 Secp256k1 和 BLS 两种密钥类型
func sigTypeForKeyType(kt types.KeyType) (crypto.SigType, error) {
	log.Debugf("sigTypeForKeyType: converting key type %s to signature type", kt)

	switch kt {
	case types.KTSecp256k1:
		log.Debug("sigTypeForKeyType: key type is Secp256k1")
		return crypto.SigTypeSecp256k1, nil
	case types.KTBLS:
		log.Debug("sigTypeForKeyType: key type is BLS")
		return crypto.SigTypeBLS, nil
	default:
		log.Errorf("sigTypeForKeyType: unsupported key type: %s", kt)
		return crypto.SigTypeUnknown, fmt.Errorf("unsupported key type: %s", kt)
	}
}

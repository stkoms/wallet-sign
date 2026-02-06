package wallet

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"wallet-sign/internal/repository"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/crypto"
	logging "github.com/ipfs/go-log/v2"

	"wallet-sign/internal/chain/types"
)

var log = logging.Logger("wallet")

// WalletSign 使用指定地址的私钥签名消息
// 从数据库中查找密钥并执行签名操作
func WalletSign(store *repository.Store, addr address.Address, msg []byte) (*crypto.Signature, error) {
	log.Infof("WalletSign: signing message for address %s", addr.String())

	res, err := store.GetWalletKey(addr.String())
	if err != nil {
		log.Errorf("WalletSign: failed to get key for %s: %v", addr.String(), err)
		return nil, fmt.Errorf("getting key for %s: %w", addr.String(), err)
	}

	var ki types.KeyInfo
	if err = json.Unmarshal(res.EncryptedKey, &ki); err != nil {
		log.Errorf("WalletSign: failed to unmarshal key for %s: %v", addr.String(), err)
		return nil, fmt.Errorf("unmarshaling key: %w", err)
	}

	sigType, err := sigTypeForKeyType(ki.Type)
	if err != nil {
		log.Errorf("WalletSign: invalid key type for %s: %v", addr.String(), err)
		return nil, err
	}

	sigBytes, err := SignBytes(msg, ki.PrivateKey, sigType)
	if err != nil {
		log.Errorf("WalletSign: failed to sign bytes for %s: %v", addr.String(), err)
		return nil, err
	}

	log.Infof("WalletSign: successfully signed message for address %s", addr.String())
	return &crypto.Signature{
		Type: sigType,
		Data: sigBytes,
	}, nil
}

// WalletImport 导入密钥到钱包
// 从密钥信息派生地址
func WalletImport(ki *types.KeyInfo) (address.Address, error) {
	log.Infof("WalletImport: importing key of type %s", ki.Type)

	sigType, err := sigTypeForKeyType(ki.Type)
	if err != nil {
		log.Errorf("WalletImport: invalid key type %s: %v", ki.Type, err)
		return address.Undef, err
	}

	addr, err := PrivateKeyToAddress(ki.PrivateKey, sigType)
	if err != nil {
		log.Errorf("WalletImport: failed to derive address: %v", err)
		return address.Undef, fmt.Errorf("failed to make key: %w", err)
	}

	log.Infof("WalletImport: successfully imported key, address: %s", addr.String())
	return addr, nil
}

// WalletNew 创建新的钱包密钥
// 根据指定的密钥类型生成新密钥
func WalletNew(typ types.KeyType) (*types.KeyInfo, address.Address, error) {
	log.Infof("WalletNew: generating new key of type %s", typ)

	var privKey []byte
	var err error
	var sigType crypto.SigType

	switch typ {
	case types.KTSecp256k1:
		privKey, err = GenerateKey()
		if err != nil {
			log.Errorf("WalletNew: failed to generate secp256k1 key: %v", err)
			return nil, address.Undef, fmt.Errorf("failed to generate secp256k1 key: %w", err)
		}
		sigType = crypto.SigTypeSecp256k1

	case types.KTBLS:
		seed := make([]byte, 32)
		if _, err := io.ReadFull(rand.Reader, seed); err != nil {
			log.Errorf("WalletNew: failed to generate random seed: %v", err)
			return nil, address.Undef, fmt.Errorf("failed to generate random seed: %w", err)
		}
		privKey, err = BLSGeneratePrivateKeyWithSeed(seed)
		if err != nil {
			log.Errorf("WalletNew: failed to generate BLS key: %v", err)
			return nil, address.Undef, fmt.Errorf("failed to generate BLS key: %w", err)
		}
		sigType = crypto.SigTypeBLS

	default:
		log.Errorf("WalletNew: unsupported key type: %s", typ)
		return nil, address.Undef, fmt.Errorf("unsupported key type: %s", typ)
	}

	addr, err := PrivateKeyToAddress(privKey, sigType)
	if err != nil {
		log.Errorf("WalletNew: failed to derive address: %v", err)
		return nil, address.Undef, fmt.Errorf("failed to derive address: %w", err)
	}

	ki := &types.KeyInfo{
		Type:       typ,
		PrivateKey: privKey,
	}

	log.Infof("WalletNew: successfully generated new key, address: %s", addr.String())
	return ki, addr, nil
}

// WalletHas 检查数据库中是否有指定地址的密钥
func WalletHas(store *repository.Store, addr address.Address) (bool, error) {
	log.Debugf("WalletHas: checking if key exists for address %s", addr.String())

	_, err := store.GetWalletKey(addr.String())
	if err != nil {
		log.Debugf("WalletHas: key not found for address %s", addr.String())
		return false, nil
	}

	log.Debugf("WalletHas: key found for address %s", addr.String())
	return true, nil
}

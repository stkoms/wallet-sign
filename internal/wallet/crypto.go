package wallet

import (
	"fmt"

	"github.com/filecoin-project/go-address"
	fcrypto "github.com/filecoin-project/go-crypto"
	"github.com/filecoin-project/go-state-types/crypto"
	"golang.org/x/crypto/blake2b"
)

// SignBytes signs data with a private key using the specified signature type.
// Supports both secp256k1 and BLS signing using pure Go implementations.
func SignBytes(data []byte, privKey []byte, sigType crypto.SigType) ([]byte, error) {
	log.Debugf("SignBytes: signing data with signature type %d", sigType)

	switch sigType {
	case crypto.SigTypeSecp256k1:
		digest := blake2b.Sum256(data)
		sig, err := fcrypto.Sign(privKey, digest[:])
		if err != nil {
			log.Errorf("SignBytes: secp256k1 signing failed: %v", err)
			return nil, err
		}
		log.Debug("SignBytes: secp256k1 signing successful")
		return sig, nil

	case crypto.SigTypeBLS:
		// Use pure Go BLS implementation
		sig, err := BLSSign(privKey, data)
		if err != nil {
			log.Errorf("SignBytes: BLS signing failed: %v", err)
			return nil, err
		}
		log.Debug("SignBytes: BLS signing successful")
		return sig, nil

	default:
		log.Errorf("SignBytes: unsupported signature type: %d", sigType)
		return nil, fmt.Errorf("unsupported signature type: %d", sigType)
	}
}

// PrivateKeyToAddress derives a Filecoin address from a private key.
// Supports both secp256k1 and BLS addresses using pure Go implementations.
func PrivateKeyToAddress(privKey []byte, sigType crypto.SigType) (address.Address, error) {
	log.Debugf("PrivateKeyToAddress: deriving address for signature type %d", sigType)

	var pubKey []byte
	var err error

	switch sigType {
	case crypto.SigTypeSecp256k1:
		pubKey, err = secpPublicKey(privKey)
		if err != nil {
			log.Errorf("PrivateKeyToAddress: failed to get secp256k1 public key: %v", err)
			return address.Undef, err
		}
		addr, err := address.NewSecp256k1Address(pubKey)
		if err != nil {
			log.Errorf("PrivateKeyToAddress: failed to create secp256k1 address: %v", err)
			return address.Undef, err
		}
		log.Debugf("PrivateKeyToAddress: created secp256k1 address %s", addr)
		return addr, nil

	case crypto.SigTypeBLS:
		// Use pure Go BLS implementation
		pubKey, err = BLSPrivateKeyToPublicKey(privKey)
		if err != nil {
			log.Errorf("PrivateKeyToAddress: failed to get BLS public key: %v", err)
			return address.Undef, err
		}
		addr, err := address.NewBLSAddress(pubKey)
		if err != nil {
			log.Errorf("PrivateKeyToAddress: failed to create BLS address: %v", err)
			return address.Undef, err
		}
		log.Debugf("PrivateKeyToAddress: created BLS address %s", addr)
		return addr, nil

	default:
		log.Errorf("PrivateKeyToAddress: unsupported signature type: %d", sigType)
		return address.Undef, fmt.Errorf("unsupported signature type: %d", sigType)
	}
}

func secpPublicKey(privKey []byte) (pubKey []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("secpPublicKey: panic during public key generation")
			err = fmt.Errorf("invalid secp256k1 private key")
		}
	}()
	pubKey = fcrypto.PublicKey(privKey)
	if len(pubKey) == 0 {
		log.Error("secpPublicKey: generated empty public key")
		return nil, fmt.Errorf("invalid secp256k1 private key")
	}
	log.Debug("secpPublicKey: successfully generated public key")
	return pubKey, nil
}

// GenerateKey generates a new secp256k1 private key.
// Returns 32 bytes of random data suitable for use as a secp256k1 private key.
func GenerateKey() ([]byte, error) {
	log.Debug("GenerateKey: generating new secp256k1 private key")

	privKey, err := fcrypto.GenerateKey()
	if err != nil {
		log.Errorf("GenerateKey: failed to generate secp256k1 key: %v", err)
		return nil, fmt.Errorf("failed to generate secp256k1 key: %w", err)
	}

	log.Debug("GenerateKey: successfully generated secp256k1 private key")
	return privKey, nil
}

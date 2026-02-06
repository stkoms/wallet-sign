package types

import (
	"encoding/json"
	"fmt"

	"github.com/filecoin-project/go-state-types/crypto"
)

// KeyType defines a type of key.
type KeyType string

const (
	KTSecp256k1 KeyType = "secp256k1"
	KTBLS       KeyType = "bls"
)

func (kt *KeyType) UnmarshalJSON(bb []byte) error {
	{
		var s string
		err := json.Unmarshal(bb, &s)
		if err == nil {
			*kt = KeyType(s)
			return nil
		}
	}

	{
		var b byte
		err := json.Unmarshal(bb, &b)
		if err != nil {
			return fmt.Errorf("could not unmarshal KeyType either as string nor integer: %w", err)
		}
		bst := crypto.SigType(b)

		switch bst {
		case crypto.SigTypeBLS:
			*kt = KTBLS
		case crypto.SigTypeSecp256k1:
			*kt = KTSecp256k1
		default:
			return fmt.Errorf("unsupported signature type: %d", b)
		}
	}

	return nil
}

type KeyInfo struct {
	Type       KeyType
	PrivateKey []byte
}

package wallet

import (
	"crypto/sha256"
	"fmt"
	"io"

	bls12381 "github.com/kilic/bls12-381"
	"golang.org/x/crypto/hkdf"
)

const (
	// BLSPrivateKeyBytes BLS12-381 私钥字节长度
	BLSPrivateKeyBytes = 32
	// BLSPublicKeyBytes BLS12-381 公钥字节长度
	BLSPublicKeyBytes = 48
	// BLSSignatureBytes BLS12-381 签名字节长度
	BLSSignatureBytes = 96

	// BLSDST Filecoin 中 BLS 签名的域分离标签
	BLSDST = "BLS_SIG_BLS12381G2_XMD:SHA-256_SSWU_RO_NUL_"
)

// BLSPrivateKeyToPublicKey 从私钥派生 BLS 公钥
// 这是一个纯 Go 实现，替代了 ffi.PrivateKeyPublicKey
func BLSPrivateKeyToPublicKey(privKey []byte) ([]byte, error) {
	log.Debug("BLSPrivateKeyToPublicKey: deriving BLS public key from private key")

	if len(privKey) != BLSPrivateKeyBytes {
		log.Errorf("BLSPrivateKeyToPublicKey: invalid private key length: expected %d, got %d", BLSPrivateKeyBytes, len(privKey))
		return nil, fmt.Errorf("invalid BLS private key length: expected %d, got %d", BLSPrivateKeyBytes, len(privKey))
	}

	// 创建新的 G1 点生成器
	g1 := bls12381.NewG1()

	// Filecoin 使用小端序存储私钥，但库需要大端序，因此反转字节
	privKeyReversed := make([]byte, BLSPrivateKeyBytes)
	for i := 0; i < BLSPrivateKeyBytes; i++ {
		privKeyReversed[i] = privKey[BLSPrivateKeyBytes-1-i]
	}

	// 将私钥字节转换为标量（Fr 元素）
	scalar := new(bls12381.Fr)
	scalar.FromBytes(privKeyReversed)

	// 将生成器点乘以私钥标量得到公钥
	publicKeyPoint := g1.New()
	g1.MulScalar(publicKeyPoint, g1.One(), scalar)

	// 序列化公钥点
	pubKeyBytes := g1.ToCompressed(publicKeyPoint)

	log.Debug("BLSPrivateKeyToPublicKey: successfully derived BLS public key")
	return pubKeyBytes, nil
}

// BLSSign 使用 BLS 私钥签名消息
// 这是一个纯 Go 实现，替代了 ffi.PrivateKeySign
func BLSSign(privKey []byte, message []byte) ([]byte, error) {
	log.Debugf("BLSSign: signing message of length %d bytes", len(message))

	if len(privKey) != BLSPrivateKeyBytes {
		log.Errorf("BLSSign: invalid private key length: expected %d, got %d", BLSPrivateKeyBytes, len(privKey))
		return nil, fmt.Errorf("invalid BLS private key length: expected %d, got %d", BLSPrivateKeyBytes, len(privKey))
	}

	// 创建 G2 生成器用于签名
	g2 := bls12381.NewG2()

	// 使用 Filecoin DST 将消息哈希到 G2 点
	messagePoint, err := g2.HashToCurve(message, []byte(BLSDST))
	if err != nil {
		log.Errorf("BLSSign: failed to hash message to curve: %v", err)
		return nil, fmt.Errorf("failed to hash message to curve: %w", err)
	}

	// 反转字节以进行小端序到大端序的转换
	privKeyReversed := make([]byte, BLSPrivateKeyBytes)
	for i := 0; i < BLSPrivateKeyBytes; i++ {
		privKeyReversed[i] = privKey[BLSPrivateKeyBytes-1-i]
	}

	// 将私钥字节转换为标量
	scalar := new(bls12381.Fr)
	scalar.FromBytes(privKeyReversed)

	// 签名：signature = privKey * H(message)
	signaturePoint := g2.New()
	g2.MulScalar(signaturePoint, messagePoint, scalar)

	// 序列化签名
	sigBytes := g2.ToCompressed(signaturePoint)

	log.Debug("BLSSign: successfully signed message")
	return sigBytes, nil
}

// BLSGeneratePrivateKeyWithSeed 从种子（IKM - 输入密钥材料）生成 BLS 私钥
// 这是一个纯 Go 实现，替代了 ffi.PrivateKeyGenerateWithSeed
// 使用 HKDF（基于 HMAC 的密钥派生函数）从种子派生私钥
func BLSGeneratePrivateKeyWithSeed(ikm []byte) ([]byte, error) {
	log.Debugf("BLSGeneratePrivateKeyWithSeed: generating BLS private key from seed of length %d bytes", len(ikm))

	if len(ikm) < 32 {
		log.Errorf("BLSGeneratePrivateKeyWithSeed: seed too short: got %d bytes, need at least 32", len(ikm))
		return nil, fmt.Errorf("seed must be at least 32 bytes, got %d", len(ikm))
	}

	// 使用 SHA-256 的 HKDF 从种子派生私钥
	// 这遵循 BLS 密钥生成标准
	salt := []byte("BLS-SIG-KEYGEN-SALT-")
	info := []byte("")

	// 创建 HKDF 读取器
	hkdfReader := hkdf.New(sha256.New, ikm, salt, info)

	// 生成 32 字节的私钥
	privKey := make([]byte, BLSPrivateKeyBytes)
	if _, err := io.ReadFull(hkdfReader, privKey); err != nil {
		log.Errorf("BLSGeneratePrivateKeyWithSeed: failed to derive private key: %v", err)
		return nil, fmt.Errorf("failed to derive private key: %w", err)
	}

	log.Debug("BLSGeneratePrivateKeyWithSeed: successfully generated BLS private key")
	return privKey, nil
}

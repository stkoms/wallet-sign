package repository

import (
	"encoding/json"
	"errors"
	"wallet-sign/internal/config"
	crypto2 "wallet-sign/internal/crypto"
	"wallet-sign/internal/models"

	"gorm.io/gorm"

	"wallet-sign/internal/chain/types"
)

// 加密密钥（使用 Scrypt + Argon2id 双重派生）
var encryptionKey []byte

// InitEncryptionKey 初始化加密密钥
func InitEncryptionKey() {
	seed := []byte(config.LotusConfig.Security.Seed)
	salt := crypto2.Hash256(seed)
	// 使用组合密钥派生函数（Scrypt + Argon2id）
	key, err := crypto2.GenerateEncryptKey(seed, salt)
	if err != nil {
		panic("failed to derive encryption key: " + err.Error())
	}
	encryptionKey = key
}

func (s *Store) SaveWalletKey(addr string, ki types.KeyInfo) error {
	log.Infof("SaveWalletKey: saving key for address %s, type %s", addr, ki.Type)

	raw, err := json.Marshal(ki)
	if err != nil {
		log.Errorf("SaveWalletKey: failed to marshal key info: %v", err)
		return err
	}
	enc, err := crypto2.EncryptGCM(raw, encryptionKey)
	if err != nil {
		log.Errorf("SaveWalletKey: failed to encrypt key data: %v", err)
		return err
	}

	var existing *models.WalletKey
	if err = s.DB.Where("address = ?", addr).First(&existing).Error; err == nil {
		log.Infof("SaveWalletKey: updating existing key for %s", addr)
		existing.KeyType = string(ki.Type)
		existing.EncryptedKey = enc
		if err := s.DB.Save(&existing).Error; err != nil {
			log.Errorf("SaveWalletKey: failed to update key: %v", err)
			return err
		}
		log.Infof("SaveWalletKey: successfully updated key for %s", addr)
		return nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Errorf("SaveWalletKey: database error when checking existing key: %v", err)
		return err
	}

	log.Infof("SaveWalletKey: creating new key record for %s", addr)
	item := &models.WalletKey{
		Address:      addr,
		KeyType:      string(ki.Type),
		EncryptedKey: enc,
	}
	if err := s.DB.Create(&item).Error; err != nil {
		log.Errorf("SaveWalletKey: failed to create key: %v", err)
		return err
	}
	log.Infof("SaveWalletKey: successfully saved key for %s", addr)
	return nil
}

func (s *Store) GetWalletKey(addr string) (*models.WalletKey, error) {
	log.Debugf("GetWalletKey: retrieving key for address %s", addr)

	item := &models.WalletKey{}
	if err := s.DB.Where("address = ?", addr).First(item).Error; err != nil {
		log.Warnf("GetWalletKey: key not found for address %s: %v", addr, err)
		return nil, err
	}

	dnc, err := crypto2.DecryptGCM(item.EncryptedKey, encryptionKey)
	if err != nil {
		log.Errorf("GetWalletKey: failed to decrypt key for %s: %v", addr, err)
		return nil, err
	}
	item.EncryptedKey = dnc

	log.Debugf("GetWalletKey: successfully retrieved key for %s", addr)
	return item, nil
}

func (s *Store) DeleteWalletKey(addr string) error {
	log.Infof("DeleteWalletKey: deleting key for address %s", addr)

	item := &models.WalletKey{}
	if err := s.DB.Where("address = ?", addr).First(item).Error; err != nil {
		log.Errorf("DeleteWalletKey: failed to find key for %s: %v", addr, err)
		return err
	}

	if err := s.DB.Delete(item).Error; err != nil {
		log.Errorf("DeleteWalletKey: failed to delete key for %s: %v", addr, err)
		return err
	}

	log.Infof("DeleteWalletKey: successfully deleted key for %s", addr)
	return nil
}

func (s *Store) GetAllWalletAddresses() ([]*models.WalletKey, error) {
	log.Debug("GetAllWalletAddresses: retrieving all wallet keys")

	var items []models.WalletKey
	if err := s.DB.Find(&items).Error; err != nil {
		log.Errorf("GetAllWalletAddresses: failed to query wallet keys: %v", err)
		return nil, err
	}

	log.Infof("GetAllWalletAddresses: found %d wallet keys", len(items))

	result := make([]*models.WalletKey, 0, len(items))
	for _, t := range items {
		decryptedKey, err := crypto2.DecryptGCM(t.EncryptedKey, encryptionKey)
		if err != nil {
			log.Errorf("GetAllWalletAddresses: failed to decrypt key for %s: %v", t.Address, err)
			return nil, err
		}

		wk := &models.WalletKey{
			Address:      t.Address,
			KeyType:      string(t.KeyType),
			EncryptedKey: decryptedKey,
			CreatedAt:    t.CreatedAt,
			UpdatedAt:    t.UpdatedAt,
		}
		result = append(result, wk)
	}

	log.Infof("GetAllWalletAddresses: successfully retrieved %d wallet keys", len(result))
	return result, nil
}

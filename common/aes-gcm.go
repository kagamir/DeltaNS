package common

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
)

// encryptString 使用AES-GCM加密字符串
func Encrypt(plaintext []byte, key []byte) ([]byte, error) {
	plaintextWithTime := append(GetTimeBytes(), plaintext...)

	// 创建一个AES块实例
	// The key argument should be the AES key,
	// either 16, 24, or 32 bytes to select
	// AES-128, AES-192, or AES-256.
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 创建GCM实例
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 创建一个随机的Nonce
	nonce := make([]byte, aesGCM.NonceSize()) // 12
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// 加密数据
	ciphertext := aesGCM.Seal(nil, nonce, plaintextWithTime, nil)
	nonceCiphertext := append(nonce, ciphertext...)

	return nonceCiphertext, nil
}

func Decrypt(nonceCiphertext []byte, key []byte) ([]byte, error) {
	if len(nonceCiphertext) <= 22 {
		return nil, errors.New("data too short")
	}

	// 创建一个AES块实例
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 创建GCM实例
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 解密数据
	nonceSize := aesGCM.NonceSize()
	nonce := nonceCiphertext[:nonceSize]
	ciphertext := nonceCiphertext[aesGCM.NonceSize():]

	timeWithPlaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	timestamp := timeWithPlaintext[:binary.MaxVarintLen64]
	isOnTime := CheckTime(timestamp)
	if !isOnTime {
		return nil, errors.New("time error")
	}
	plaintext := timeWithPlaintext[binary.MaxVarintLen64:]

	return plaintext, nil
}

package aliacm

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
)

// Md5 计算md5值
func Md5(s string) string {
	hash := md5.New()
	hash.Write([]byte(s))
	return hex.EncodeToString(hash.Sum(nil))
}

// HMACSHA1Encrypt 计算hmac sha1
func HMACSHA1Encrypt(encryptText, encryptKey string) (string, error) {
	hash := hmac.New(sha1.New, []byte(encryptKey))
	_, err := hash.Write([]byte(encryptText))
	if err != nil {
		return "", err
	}
	digest := hash.Sum(nil)
	return base64.StdEncoding.EncodeToString(digest), nil
}

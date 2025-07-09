package encrypt

import (
	"encoding/base64"
	"errors"

	"github.com/fernet/fernet-go"
)

var (
	// 전역 변수로 키를 저장합니다.
	fernetKey *fernet.Key
)

// InjectFernetKey 함수는 애플리케이션 시작 시 한 번 호출하여 암호화 키를 설정합니다.
func InjectFernetKey(keyString string) error {
	if keyString == "" {
		return errors.New("[ERROR] fernet key cannot be empty")
	}
	parsedKey, err := parseFernetKey(keyString)
	if err != nil {
		return err
	}
	fernetKey = parsedKey
	return nil
}

// FernetEncrypt 함수는 일반 텍스트를 Fernet 토큰으로 암호화하고 Base64로 인코딩합니다.
func FernetEncrypt(plainText string) (string, error) {
	if fernetKey == nil {
		return "", errors.New("[ERROR] fernet key is not initialized")
	}
	token, err := fernet.EncryptAndSign([]byte(plainText), fernetKey)
	if err != nil {
		return "", err
	}
	// Fernet 토큰은 URL-safe Base64로 인코딩하여 저장하거나 전송하는 것이 일반적입니다.
	return base64.URLEncoding.EncodeToString(token), nil
}

// FernetDecrypt 함수는 Base64로 인코딩된 Fernet 토큰을 복호화합니다.
func FernetDecrypt(cipherText string) (string, error) {
	if fernetKey == nil {
		return "", errors.New("fernet key is not initialized")
	}
	// 먼저 Base64 문자열을 디코딩하여 원래의 Fernet 토큰(바이트 슬라이스)으로 되돌립니다.
	token, err := base64.URLEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}

	plain := fernet.VerifyAndDecrypt(token, 0, []*fernet.Key{fernetKey})
	if plain == nil {
		// 복호화 실패는 다양한 이유(잘못된 키, 변조된 토큰, 만료된 토큰 등)로 발생할 수 있습니다.
		return "", errors.New("decryption failed: invalid or tampered token")
	}
	return string(plain), nil
}

// parseFernetKey 함수는 Base64로 인코딩된 문자열 키를 fernet.Key 타입으로 파싱합니다.
func parseFernetKey(keyString string) (*fernet.Key, error) {
	// Fernet 키는 URL-safe Base64로 인코딩되어 있습니다.
	decoded, err := base64.URLEncoding.DecodeString(keyString)
	if err != nil {
		return nil, errors.New("invalid base64 in fernet key")
	}

	// Fernet 키는 반드시 32바이트여야 합니다. (128-bit signing key + 128-bit encryption key)
	if len(decoded) != 32 {
		return nil, errors.New("invalid fernet key length: must be 32 bytes")
	}

	// 디코딩된 바이트를 사용하여 fernet.Key를 생성합니다.
	var key fernet.Key
	copy(key[:], decoded)

	return &key, nil
}

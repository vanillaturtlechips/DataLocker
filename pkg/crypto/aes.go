// Package crypto provides cryptographic utilities for DataLocker application.
// It implements AES-256-GCM encryption/decryption with PBKDF2 key derivation.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

// 암호화 관련 상수
const (
	// AES-256 키 크기 (32 바이트)
	KeySize = 32

	// GCM Nonce 크기 (12 바이트)
	NonceSize = 12

	// Salt 크기 (32 바이트)
	SaltSize = 32

	// PBKDF2 반복 횟수
	PBKDF2Iterations = 100000

	// 파일 청크 크기 (1MB)
	ChunkSize = 1024 * 1024

	// 청크 크기 정보를 저장할 바이트 수
	ChunkSizeBytes = 4

	// 비트 시프트 상수
	BitShift8  = 8
	BitShift16 = 16
	BitShift24 = 24

	// 최대 청크 크기 (4GB)
	MaxChunkSize = 1<<32 - 1
)

// CryptoEngine AES 암복호화 엔진
type CryptoEngine struct {
	// 추후 확장을 위한 구조체
}

// NewCryptoEngine 새로운 암호화 엔진을 생성합니다
func NewCryptoEngine() *CryptoEngine {
	return &CryptoEngine{}
}

// EncryptedData 암호화된 데이터 구조체
type EncryptedData struct {
	Salt       []byte `json:"salt"`       // PBKDF2 Salt
	Nonce      []byte `json:"nonce"`      // GCM Nonce
	Ciphertext []byte `json:"ciphertext"` // 암호화된 데이터
}

// DeriveKey PBKDF2를 사용하여 패스워드에서 키를 유도합니다
func (ce *CryptoEngine) DeriveKey(password string, salt []byte) []byte {
	return pbkdf2.Key([]byte(password), salt, PBKDF2Iterations, KeySize, sha256.New)
}

// GenerateSalt 새로운 랜덤 Salt를 생성합니다
func (ce *CryptoEngine) GenerateSalt() ([]byte, error) {
	salt := make([]byte, SaltSize)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, fmt.Errorf("salt 생성 실패: %w", err)
	}
	return salt, nil
}

// GenerateNonce 새로운 랜덤 Nonce를 생성합니다
func (ce *CryptoEngine) GenerateNonce() ([]byte, error) {
	nonce := make([]byte, NonceSize)
	_, err := rand.Read(nonce)
	if err != nil {
		return nil, fmt.Errorf("nonce 생성 실패: %w", err)
	}
	return nonce, nil
}

// Encrypt 데이터를 AES-256-GCM으로 암호화합니다
func (ce *CryptoEngine) Encrypt(plaintext []byte, password string) (*EncryptedData, error) {
	if len(plaintext) == 0 {
		return nil, errors.New("빈 데이터는 암호화할 수 없습니다")
	}

	if password == "" {
		return nil, errors.New("패스워드가 필요합니다")
	}

	// Salt 생성
	salt, err := ce.GenerateSalt()
	if err != nil {
		return nil, fmt.Errorf("salt 생성 실패: %w", err)
	}

	// 키 유도
	key := ce.DeriveKey(password, salt)

	// AES 블록 암호 생성
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("AES 암호 생성 실패: %w", err)
	}

	// GCM 모드 생성
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("GCM 모드 생성 실패: %w", err)
	}

	// Nonce 생성
	nonce, err := ce.GenerateNonce()
	if err != nil {
		return nil, fmt.Errorf("nonce 생성 실패: %w", err)
	}

	// 암호화 수행
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	return &EncryptedData{
		Salt:       salt,
		Nonce:      nonce,
		Ciphertext: ciphertext,
	}, nil
}

// Decrypt AES-256-GCM으로 암호화된 데이터를 복호화합니다
func (ce *CryptoEngine) Decrypt(encData *EncryptedData, password string) ([]byte, error) {
	if encData == nil {
		return nil, errors.New("암호화된 데이터가 없습니다")
	}

	if password == "" {
		return nil, errors.New("패스워드가 필요합니다")
	}

	// 데이터 유효성 검사
	if len(encData.Salt) != SaltSize {
		return nil, fmt.Errorf("잘못된 salt 크기: %d (예상: %d)", len(encData.Salt), SaltSize)
	}

	if len(encData.Nonce) != NonceSize {
		return nil, fmt.Errorf("잘못된 nonce 크기: %d (예상: %d)", len(encData.Nonce), NonceSize)
	}

	if len(encData.Ciphertext) == 0 {
		return nil, errors.New("암호화된 데이터가 비어있습니다")
	}

	// 키 유도
	key := ce.DeriveKey(password, encData.Salt)

	// AES 블록 암호 생성
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("AES 암호 생성 실패: %w", err)
	}

	// GCM 모드 생성
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("GCM 모드 생성 실패: %w", err)
	}

	// 복호화 수행
	plaintext, err := gcm.Open(nil, encData.Nonce, encData.Ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("복호화 실패 (잘못된 패스워드 또는 손상된 데이터): %w", err)
	}

	return plaintext, nil
}

// EncryptStream 스트림 방식으로 대용량 데이터를 암호화합니다
func (ce *CryptoEngine) EncryptStream(reader io.Reader, writer io.Writer, password string) error {
	if password == "" {
		return errors.New("패스워드가 필요합니다")
	}

	// Salt 생성 및 저장
	salt, err := ce.GenerateSalt()
	if err != nil {
		return fmt.Errorf("salt 생성 실패: %w", err)
	}

	// Salt를 파일 시작 부분에 저장
	if _, writeErr := writer.Write(salt); writeErr != nil {
		return fmt.Errorf("salt 저장 실패: %w", writeErr)
	}

	// 키 유도
	key := ce.DeriveKey(password, salt)

	// AES 블록 암호 생성
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("AES 암호 생성 실패: %w", err)
	}

	// GCM 모드 생성
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("GCM 모드 생성 실패: %w", err)
	}

	// 청크 단위로 암호화
	buffer := make([]byte, ChunkSize)
	for {
		n, readErr := reader.Read(buffer)
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return fmt.Errorf("데이터 읽기 실패: %w", readErr)
		}

		// 각 청크마다 새로운 nonce 생성
		nonce, nonceErr := ce.GenerateNonce()
		if nonceErr != nil {
			return fmt.Errorf("nonce 생성 실패: %w", nonceErr)
		}

		// Nonce 저장
		if _, writeErr := writer.Write(nonce); writeErr != nil {
			return fmt.Errorf("nonce 저장 실패: %w", writeErr)
		}

		// 청크 암호화
		chunk := buffer[:n]
		ciphertext := gcm.Seal(nil, nonce, chunk, nil)

		// 암호화된 청크 크기 검증 및 저장
		ciphertextLen := len(ciphertext)
		if ciphertextLen > MaxChunkSize {
			return fmt.Errorf("청크 크기가 너무 큽니다: %d bytes", ciphertextLen)
		}

		chunkSize := uint32(ciphertextLen)
		sizeBytes := []byte{
			byte(chunkSize >> BitShift24),
			byte(chunkSize >> BitShift16),
			byte(chunkSize >> BitShift8),
			byte(chunkSize),
		}
		if _, writeErr := writer.Write(sizeBytes); writeErr != nil {
			return fmt.Errorf("청크 크기 저장 실패: %w", writeErr)
		}

		// 암호화된 데이터 저장
		if _, writeErr := writer.Write(ciphertext); writeErr != nil {
			return fmt.Errorf("암호화된 데이터 저장 실패: %w", writeErr)
		}
	}

	return nil
}

// DecryptStream 스트림 방식으로 대용량 데이터를 복호화합니다
func (ce *CryptoEngine) DecryptStream(reader io.Reader, writer io.Writer, password string) error {
	if password == "" {
		return errors.New("패스워드가 필요합니다")
	}

	// Salt 읽기
	salt := make([]byte, SaltSize)
	if _, err := io.ReadFull(reader, salt); err != nil {
		return fmt.Errorf("salt 읽기 실패: %w", err)
	}

	// 키 유도
	key := ce.DeriveKey(password, salt)

	// AES 블록 암호 생성
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("AES 암호 생성 실패: %w", err)
	}

	// GCM 모드 생성
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("GCM 모드 생성 실패: %w", err)
	}

	// 청크 단위로 복호화
	for {
		// Nonce 읽기
		nonce := make([]byte, NonceSize)
		n, readErr := reader.Read(nonce)
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return fmt.Errorf("nonce 읽기 실패: %w", readErr)
		}
		if n != NonceSize {
			return fmt.Errorf("잘못된 nonce 크기: %d", n)
		}

		// 청크 크기 읽기
		sizeBytes := make([]byte, ChunkSizeBytes)
		if _, readFullErr := io.ReadFull(reader, sizeBytes); readFullErr != nil {
			return fmt.Errorf("청크 크기 읽기 실패: %w", readFullErr)
		}

		chunkSize := uint32(sizeBytes[0])<<BitShift24 |
			uint32(sizeBytes[1])<<BitShift16 |
			uint32(sizeBytes[2])<<BitShift8 |
			uint32(sizeBytes[3])

		// 암호화된 데이터 읽기
		ciphertext := make([]byte, chunkSize)
		if _, readFullErr := io.ReadFull(reader, ciphertext); readFullErr != nil {
			return fmt.Errorf("암호화된 데이터 읽기 실패: %w", readFullErr)
		}

		// 복호화
		plaintext, decryptErr := gcm.Open(nil, nonce, ciphertext, nil)
		if decryptErr != nil {
			return fmt.Errorf("복호화 실패: %w", decryptErr)
		}

		// 복호화된 데이터 저장
		if _, writeErr := writer.Write(plaintext); writeErr != nil {
			return fmt.Errorf("복호화된 데이터 저장 실패: %w", writeErr)
		}
	}

	return nil
}

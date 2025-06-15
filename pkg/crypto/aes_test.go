package crypto

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 테스트용 상수
const (
	TestPassword      = "testpassword"
	TestData          = "Hello, DataLocker!"
	LongTestData      = "DataLocker는 안전한 파일 암호화 솔루션입니다. "
	BenchmarkPassword = "benchmarkpassword"
	StreamPassword    = "streampassword"
	LongDataRepeat    = 100
	BenchmarkRepeat   = 100
	StreamTestRepeat  = 1000
)

// 테스트용 데이터
var (
	testBinaryData = []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD}
	testSalt       = []byte("testsalt12345678901234567890123")
)

func TestNewCryptoEngine(t *testing.T) {
	engine := NewCryptoEngine()
	assert.NotNil(t, engine)
}

func TestGenerateSalt(t *testing.T) {
	engine := NewCryptoEngine()

	salt, err := engine.GenerateSalt()
	require.NoError(t, err)
	assert.Len(t, salt, SaltSize)

	// 두 번 생성한 salt는 달라야 함
	salt2, err := engine.GenerateSalt()
	require.NoError(t, err)
	assert.NotEqual(t, salt, salt2)
}

func TestGenerateNonce(t *testing.T) {
	engine := NewCryptoEngine()

	nonce, err := engine.GenerateNonce()
	require.NoError(t, err)
	assert.Len(t, nonce, NonceSize)

	// 두 번 생성한 nonce는 달라야 함
	nonce2, err := engine.GenerateNonce()
	require.NoError(t, err)
	assert.NotEqual(t, nonce, nonce2)
}

func TestDeriveKey(t *testing.T) {
	engine := NewCryptoEngine()

	key := engine.DeriveKey(TestPassword, testSalt)
	assert.Len(t, key, KeySize)

	// 같은 패스워드와 salt로는 같은 키가 나와야 함
	key2 := engine.DeriveKey(TestPassword, testSalt)
	assert.Equal(t, key, key2)

	// 다른 패스워드로는 다른 키가 나와야 함
	key3 := engine.DeriveKey("differentpassword", testSalt)
	assert.NotEqual(t, key, key3)
}

func TestEncryptDecrypt_Success(t *testing.T) {
	engine := NewCryptoEngine()

	testCases := []struct {
		name     string
		data     []byte
		password string
	}{
		{
			name:     "작은 텍스트 데이터",
			data:     []byte(TestData),
			password: "mypassword123",
		},
		{
			name:     "빈 문자열이 아닌 작은 데이터",
			data:     []byte("a"),
			password: "pass",
		},
		{
			name:     "긴 텍스트 데이터",
			data:     []byte(strings.Repeat(LongTestData, LongDataRepeat)),
			password: "verysecurepassword!@#$",
		},
		{
			name:     "바이너리 데이터",
			data:     testBinaryData,
			password: "binarytest",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 암호화
			encData, err := engine.Encrypt(tc.data, tc.password)
			require.NoError(t, err)
			require.NotNil(t, encData)

			// 암호화된 데이터 검증
			assert.Len(t, encData.Salt, SaltSize)
			assert.Len(t, encData.Nonce, NonceSize)
			assert.Greater(t, len(encData.Ciphertext), 0)
			assert.NotEqual(t, tc.data, encData.Ciphertext)

			// 복호화
			decrypted, err := engine.Decrypt(encData, tc.password)
			require.NoError(t, err)
			assert.Equal(t, tc.data, decrypted)
		})
	}
}

func TestEncrypt_ErrorCases(t *testing.T) {
	engine := NewCryptoEngine()

	testCases := []struct {
		name     string
		data     []byte
		password string
		wantErr  string
	}{
		{
			name:     "빈 데이터",
			data:     []byte{},
			password: "password",
			wantErr:  "빈 데이터는 암호화할 수 없습니다",
		},
		{
			name:     "nil 데이터",
			data:     nil,
			password: "password",
			wantErr:  "빈 데이터는 암호화할 수 없습니다",
		},
		{
			name:     "빈 패스워드",
			data:     []byte("test"),
			password: "",
			wantErr:  "패스워드가 필요합니다",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := engine.Encrypt(tc.data, tc.password)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestDecrypt_ErrorCases(t *testing.T) {
	engine := NewCryptoEngine()

	testCases := []struct {
		name    string
		data    *EncryptedData
		passwd  string
		wantErr string
	}{
		{
			name:    "nil 데이터",
			data:    nil,
			passwd:  "password",
			wantErr: "암호화된 데이터가 없습니다",
		},
		{
			name: "잘못된 salt 크기",
			data: &EncryptedData{
				Salt:       []byte("short"),
				Nonce:      make([]byte, NonceSize),
				Ciphertext: []byte("test"),
			},
			passwd:  "password",
			wantErr: "잘못된 salt 크기",
		},
		{
			name: "잘못된 nonce 크기",
			data: &EncryptedData{
				Salt:       make([]byte, SaltSize),
				Nonce:      []byte("short"),
				Ciphertext: []byte("test"),
			},
			passwd:  "password",
			wantErr: "잘못된 nonce 크기",
		},
		{
			name: "빈 ciphertext",
			data: &EncryptedData{
				Salt:       make([]byte, SaltSize),
				Nonce:      make([]byte, NonceSize),
				Ciphertext: []byte{},
			},
			passwd:  "password",
			wantErr: "암호화된 데이터가 비어있습니다",
		},
		{
			name: "빈 패스워드",
			data: &EncryptedData{
				Salt:       make([]byte, SaltSize),
				Nonce:      make([]byte, NonceSize),
				Ciphertext: []byte("test"),
			},
			passwd:  "",
			wantErr: "패스워드가 필요합니다",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := engine.Decrypt(tc.data, tc.passwd)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestDecrypt_WrongPassword(t *testing.T) {
	engine := NewCryptoEngine()

	// 정상 암호화
	data := []byte("test data")
	encData, err := engine.Encrypt(data, "correctpassword")
	require.NoError(t, err)

	// 잘못된 패스워드로 복호화 시도
	_, err = engine.Decrypt(encData, "wrongpassword")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "복호화 실패")
}

func TestEncryptStream_Success(t *testing.T) {
	engine := NewCryptoEngine()

	testData := []byte(strings.Repeat("DataLocker Stream Test ", StreamTestRepeat))

	// 암호화
	var encryptedBuf bytes.Buffer
	reader := bytes.NewReader(testData)

	err := engine.EncryptStream(reader, &encryptedBuf, StreamPassword)
	require.NoError(t, err)
	assert.Greater(t, encryptedBuf.Len(), len(testData))

	// 복호화
	var decryptedBuf bytes.Buffer
	encryptedReader := bytes.NewReader(encryptedBuf.Bytes())

	err = engine.DecryptStream(encryptedReader, &decryptedBuf, StreamPassword)
	require.NoError(t, err)
	assert.Equal(t, testData, decryptedBuf.Bytes())
}

func TestEncryptStream_EmptyData(t *testing.T) {
	engine := NewCryptoEngine()

	var encryptedBuf bytes.Buffer
	reader := bytes.NewReader([]byte{})

	err := engine.EncryptStream(reader, &encryptedBuf, "password")
	require.NoError(t, err)

	// Salt만 저장되어야 함
	assert.Equal(t, SaltSize, encryptedBuf.Len())
}

func TestEncryptStream_ErrorCases(t *testing.T) {
	engine := NewCryptoEngine()

	var buf bytes.Buffer
	reader := bytes.NewReader([]byte("test"))

	err := engine.EncryptStream(reader, &buf, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "패스워드가 필요합니다")
}

func TestDecryptStream_ErrorCases(t *testing.T) {
	engine := NewCryptoEngine()

	testCases := []struct {
		name    string
		data    []byte
		passwd  string
		wantErr string
	}{
		{
			name:    "빈 패스워드",
			data:    []byte("test"),
			passwd:  "",
			wantErr: "패스워드가 필요합니다",
		},
		{
			name:    "짧은 데이터 (salt 없음)",
			data:    []byte("short"),
			passwd:  "password",
			wantErr: "salt 읽기 실패",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			reader := bytes.NewReader(tc.data)

			err := engine.DecryptStream(reader, &buf, tc.passwd)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestEncryptDecryptStream_WrongPassword(t *testing.T) {
	engine := NewCryptoEngine()

	testData := []byte("test stream data")

	// 정상 암호화
	var encryptedBuf bytes.Buffer
	reader := bytes.NewReader(testData)
	err := engine.EncryptStream(reader, &encryptedBuf, "correctpassword")
	require.NoError(t, err)

	// 잘못된 패스워드로 복호화
	var decryptedBuf bytes.Buffer
	encryptedReader := bytes.NewReader(encryptedBuf.Bytes())

	err = engine.DecryptStream(encryptedReader, &decryptedBuf, "wrongpassword")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "복호화 실패")
}

// 벤치마크 테스트
func BenchmarkEncrypt(b *testing.B) {
	engine := NewCryptoEngine()
	data := []byte(strings.Repeat("benchmark test data ", BenchmarkRepeat))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.Encrypt(data, BenchmarkPassword)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecrypt(b *testing.B) {
	engine := NewCryptoEngine()
	data := []byte(strings.Repeat("benchmark test data ", BenchmarkRepeat))

	encData, err := engine.Encrypt(data, BenchmarkPassword)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.Decrypt(encData, BenchmarkPassword)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncryptStream(b *testing.B) {
	engine := NewCryptoEngine()
	data := []byte(strings.Repeat("stream benchmark test data ", StreamTestRepeat))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		reader := bytes.NewReader(data)

		err := engine.EncryptStream(reader, &buf, BenchmarkPassword)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecryptStream(b *testing.B) {
	engine := NewCryptoEngine()
	data := []byte(strings.Repeat("stream benchmark test data ", StreamTestRepeat))

	// 미리 암호화된 데이터 준비
	var encryptedBuf bytes.Buffer
	reader := bytes.NewReader(data)
	err := engine.EncryptStream(reader, &encryptedBuf, BenchmarkPassword)
	if err != nil {
		b.Fatal(err)
	}

	encryptedData := encryptedBuf.Bytes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		reader := bytes.NewReader(encryptedData)

		err := engine.DecryptStream(reader, &buf, BenchmarkPassword)
		if err != nil {
			b.Fatal(err)
		}
	}
}

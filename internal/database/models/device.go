package models

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/scrypt"
)

type Device struct {
	Id             int
	Name           string
	SN             string
	Paired         time.Time
	Hardware       *string
	Firmware       *string
	TargetFirmware *string
	KeyCycle       int
	MakerspaceId   int
}

func (d Device) LogEntity() LogEntity {
	return LogEntity{
		Id:    d.Id,
		Label: d.Name,
	}
}

func (dev Device) CredentialsMatch(SN string, key string) (bool, error) {
	if SN != dev.SN {
		return false, nil
	}
	keyToMatch, err := dev.GenerateKey()
	if err != nil {
		return false, fmt.Errorf("Failed to generate key: %w", err)
	}
	return keyToMatch == key, nil
}

// PKCS7Padding adds padding to the plaintext so its length is a multiple of the block size
// needed for compatibility with original implementation
func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - (len(ciphertext) % blockSize)
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func (dev Device) GenerateKey() (string, error) {
	// 1. Get environment variable or fallback
	serverApiPass := os.Getenv("SERVER_API_PASSWORD")
	if serverApiPass == "" {
		serverApiPass = "unsecure_server_password"
	}

	// 2. Node.js scryptSync default parameters: N=16384, r=8, p=1
	salt := []byte("makerspace-salt¯_(ツ)_/¯")
	serverKey, err := scrypt.Key([]byte(serverApiPass), salt, 16384, 8, 1, 24)
	if err != nil {
		return "", fmt.Errorf("scrypt error: %w", err)
	}

	// 3. Generate plain text string
	plainTextStr := fmt.Sprintf("shlug:%s:%d", dev.SN, dev.KeyCycle)
	plainText := []byte(plainTextStr)
	// 4. Generate IV (SHA-256 hash of ISO string, sliced to 16 bytes)
	// Node's toISOString() uses UTC with 'Z' suffix (e.g., 2026-07-08T12:46:00.000Z)
	isoTimeStr := dev.Paired.UTC().Format("2006-01-02T15:04:05.000Z")
	hash := sha256.Sum256([]byte(isoTimeStr))
	iv := hash[:16]

	// 5. Setup AES-192 (determined by 24-byte key length)
	block, err := aes.NewCipher(serverKey)
	if err != nil {
		return "", fmt.Errorf("cipher setup error: %w", err)
	}

	// Node's createCipheriv automatically uses PKCS7 padding
	paddedPlainText := PKCS7Padding(plainText, block.BlockSize())

	// 6. Encrypt using CBC mode
	encrypted := make([]byte, len(paddedPlainText))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(encrypted, paddedPlainText)

	// Return as hex string
	return hex.EncodeToString(encrypted), nil
}

type AccessDeviceFlags struct {
	LockWhenIdle      bool `json:"lock_when_idle"`
	RestartWhenUnused bool `json:"restart_when_unused"`
	Welcoming         bool `json:"welcoming"`
}

func (f *AccessDeviceFlags) Scan(value any) error {
	if value == nil {
		*f = AccessDeviceFlags{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed for AccessDeviceFlags")
	}

	return json.Unmarshal(bytes, f)
}

func (f AccessDeviceFlags) Value() (driver.Value, error) {
	return json.Marshal(f)
}

type AccessComponentType int

const (
	Legacy         AccessComponentType = 0b000
	Switch1Channel AccessComponentType = 0b001
	Switch2Channel AccessComponentType = 0b010
	Switch3Channel AccessComponentType = 0b011
	Switch4Channel AccessComponentType = 0b100
	NonSwitching   AccessComponentType = 0b101
	InternalHMI    AccessComponentType = 0b110
	Communicative  AccessComponentType = 0b111
)

type AccessComponent struct {
	SN       string              `json:"SN"`
	Type     AccessComponentType `json:"type"`
	Children []AccessComponent   `json:"children"`
}

type AccessDeployment struct {
	SN         string            `json:"SN"`
	Components []AccessComponent `json:"components"`
}

func (d *AccessDeployment) Scan(value any) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed for AccessDeployment")
	}

	return json.Unmarshal(bytes, d)
}

func (d AccessDeployment) Value() (driver.Value, error) {
	return json.Marshal(d)
}

type AccessDeviceInputMode string

var (
	AccessDeviceInputMode_Insert      AccessDeviceInputMode = "INSERT"
	AccessDeviceInputMode_TempPresent AccessDeviceInputMode = "TEMP_PRESENT"
	AccessDeviceInputMode_TempRemove  AccessDeviceInputMode = "TEMP_REMOVE"
	AccessDeviceInputMode_Toggle      AccessDeviceInputMode = "TOGGLE"
)

type AccessDevice struct {
	Device
	Channels           int
	TempDuration       int
	CurrentCardTag     string
	LastStatus         *time.Time
	SessionStart       *time.Time
	Flags              AccessDeviceFlags
	SealedDeployment   *AccessDeployment
	ReportedDeployment *AccessDeployment
}

type DispenserError string

const (
	CardStuck  DispenserError = "CARD_STUCK"
	OutOfCards DispenserError = "OUT_OF_CARDS"
)

type Dispenser struct {
	DeviceId  int
	CardsLeft int
	Error     *DispenserError
}

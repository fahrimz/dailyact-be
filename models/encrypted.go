package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// EncryptedString is a type that automatically handles encryption and decryption
// of string values when interacting with the database
type EncryptedString string

// Value encrypts the string before storing it in the database
func (es *EncryptedString) Value() (driver.Value, error) {
	if es == nil {
		return nil, nil
	}

	// Get encryption service
	encryptionService, err := NewEncryptionService()
	if err != nil {
		return nil, err
	}

	encrypted, err := encryptionService.Encrypt(string(*es))
	if err != nil {
		return nil, err
	}

	return encrypted, nil
}

// Scan decrypts the string when reading from the database
func (es *EncryptedString) Scan(value interface{}) error {
	if value == nil {
		*es = ""
		return nil
	}

	// Get encryption service
	encryptionService, err := NewEncryptionService()
	if err != nil {
		return err
	}

	var byteValue []byte
	switch v := value.(type) {
	case []byte:
		byteValue = v
	case string:
		byteValue = []byte(v)
	default:
		return errors.New("unsupported type for encrypted string")
	}

	decrypted, err := encryptionService.Decrypt(string(byteValue))
	if err != nil {
		return err
	}

	*es = EncryptedString(decrypted)
	return nil
}

// MarshalJSON returns the string as a JSON value
// This ensures that the encrypted string is properly sent as a regular string in JSON responses
func (es EncryptedString) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(es))
}

// String returns the decrypted string
func (es EncryptedString) String() string {
	return string(es)
}

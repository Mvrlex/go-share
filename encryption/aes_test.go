package encryption

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestEncryption(t *testing.T) {
	aes := Aes{}

	hash, encryptReader, err := aes.Encrypt(strings.NewReader("value"), "password", "salt")
	if err != nil {
		t.Error(err)
	}
	encrypted := make([]byte, 5)
	_, err = encryptReader.Read(encrypted)
	if err != nil && err != io.EOF {
		t.Error(err)
	}
	t.Logf("Value to encrypt: \"value\"\nEncrypted result: %s\nHash used:%s", encrypted, hash)
	if string(encrypted) == "value" {
		t.Errorf("encrypted value matches input")
	}

	decryptReader, err := aes.Decrypt(bytes.NewReader(encrypted), "password", "salt")
	if err != nil {
		t.Error(err)
	}
	decrypted := make([]byte, 5)
	_, err = decryptReader.Read(decrypted)
	if err != nil && err != io.EOF {
		t.Error(err)
	}
	t.Logf("Decrypted result: %s", decrypted)
	if string(decrypted) != "value" {
		t.Errorf("decrypted value does not match input, \"%s\" != \"value\"", string(decrypted))
	}

}

func TestHashVerification(t *testing.T) {
	aes := Aes{}

	hash, _, err := aes.Encrypt(strings.NewReader("value"), "password", "salt")
	if err != nil {
		t.Error(err)
	}
	if !aes.Verify("password", "salt", hash) {
		t.Errorf("Generated hashes from the same values did not match")
	}
	if aes.Verify("smashword", "salt", hash) {
		t.Errorf("Generated hashes from different passwords did match")
	}
	if aes.Verify("password", "pepper", hash) {
		t.Errorf("Generated hashes from different salts did match")
	}

}

func TestPasswordLengthValidation(t *testing.T) {
	aes := Aes{}

	_, _, err := aes.Encrypt(strings.NewReader(""), "123", "")
	if !errors.Is(err, PasswordTooShortError) {
		t.Errorf("Passwords with length below 4 should not be allowed")
	}

	_, _, err = aes.Encrypt(strings.NewReader(""), "1234", "")
	if err != nil {
		t.Errorf("Passwords with length 4 should be allowed")
	}

	_, _, err = aes.Encrypt(strings.NewReader(""), "123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789", "")
	if !errors.Is(err, PasswordTooLongError) {
		t.Errorf("Passwords with length above 128 should not be allowed")
	}

	_, _, err = aes.Encrypt(strings.NewReader(""), "12345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678", "")
	if err != nil {
		t.Errorf("Passwords with length 128 should be allowed")
	}

}

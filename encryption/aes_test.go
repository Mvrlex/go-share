package encryption

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestEncryption(t *testing.T) {
	aes := Aes{}

	hash, encryptReader := aes.Encrypt(strings.NewReader("value"), "password", "salt")
	encrypted := make([]byte, 5)
	_, err := encryptReader.Read(encrypted)
	if err != nil && err != io.EOF {
		t.Error(err)
	}
	t.Logf("Value to encrypt: \"value\"\nEncrypted result: %s\nHash used:%s", encrypted, hash)
	if string(encrypted) == "value" {
		t.Errorf("encrypted value matches input")
	}

	decryptReader := aes.Decrypt(bytes.NewReader(encrypted), "password", "salt")
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

	hash, _ := aes.Encrypt(strings.NewReader("value"), "password", "salt")
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

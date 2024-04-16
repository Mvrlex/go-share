package encryption

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"golang.org/x/crypto/argon2"
	"io"
)

var (
	PasswordTooShortError = errors.New("provided password is too short")
	PasswordTooLongError  = errors.New("provided password is too long")
)

type Aes struct {
}

func (a *Aes) Encrypt(value io.Reader, password, salt string) (pwHash string, encrypted io.Reader, err error) {
	genHash, err := hash([]byte(password), []byte(salt))
	if err != nil {
		return "", nil, err
	}
	return string(genHash), crypt(value, genHash), nil
}

func (a *Aes) Decrypt(value io.Reader, password, salt string) (io.Reader, error) {
	genHash, err := hash([]byte(password), []byte(salt))
	if err != nil {
		return nil, err
	}
	return crypt(value, genHash), nil
}

func (a *Aes) Verify(password, salt, pwHash string) bool {
	genHash, err := hash([]byte(password), []byte(salt))
	if err != nil {
		return false
	}
	return bytes.Equal(genHash, []byte(pwHash))
}

func hash(password, salt []byte) ([]byte, error) {
	if err := validatePassword(password); err != nil {
		return nil, err
	}
	return argon2.IDKey(password, salt, 1, 64*1024, 4, 32), nil
}

func validatePassword(password []byte) error {
	if len(password) < 4 {
		return PasswordTooShortError
	}
	if len(password) > 128 {
		return PasswordTooLongError
	}
	return nil
}

// crypt both encrypts and decrypts a value using the given password and salt.
func crypt(value io.Reader, hash []byte) io.Reader {
	block, err := aes.NewCipher(hash)
	if err != nil {
		panic(err) // this can only ever be an implementation error
	}
	// Design note: Go recommends using a HMAC to ensure data integrity, but I think
	// that it is unnecessary in this context, as we save the file to our own local
	// filesystem. There is pretty much no way anyone can tamper with the file, and even
	// if they could, they would only be able to corrupt the encrypted file.
	// FIXME add file corruption verification via a sha1 hash?
	// The key for the cipher is already unique enough, an additional IV would not provide
	// any additional protection, that's why we simply provide a static IV.
	stream := cipher.NewCTR(block, []byte("higgledypiggledy"))
	encrypted, writer := io.Pipe()
	go func() {
		defer writer.Close()
		buffIn := make([]byte, 1024)
		buffOut := make([]byte, 1024)
		for {
			n, readErr := value.Read(buffIn)

			stream.XORKeyStream(buffOut, buffIn[:n])
			_, err = writer.Write(buffOut[:n])
			if err != nil {
				writer.CloseWithError(err)
				break
			}

			if readErr == io.EOF {
				break
			} else if readErr != nil {
				writer.CloseWithError(readErr)
			}
		}
	}()
	return encrypted
}

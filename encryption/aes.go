package encryption

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"golang.org/x/crypto/argon2"
	"io"
)

type Aes struct {
}

func (a *Aes) Encrypt(value io.Reader, password, salt string) (pwHash string, encrypted io.Reader) {
	genHash := hash([]byte(password), []byte(salt))
	return string(genHash), crypt(value, genHash)
}

func (a *Aes) Decrypt(value io.Reader, password, salt string) io.Reader {
	genHash := hash([]byte(password), []byte(salt))
	return crypt(value, genHash)
}

func (a *Aes) Verify(password, salt, pwHash string) bool {
	genHash := hash([]byte(password), []byte(salt))
	return bytes.Equal(genHash, []byte(pwHash))
}

func hash(password, salt []byte) []byte {
	return argon2.IDKey(password, salt, 1, 64*1024, 4, 32)
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

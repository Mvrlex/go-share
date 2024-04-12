package encryption

import "io"

type Encryption interface {
	Encrypt(value io.Reader, password, salt string) (hash string, encrypted io.Reader)
	Decrypt(value io.Reader, password, salt string) io.Reader
	Verify(password, salt, hash string) bool
}

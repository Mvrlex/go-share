package encryption

import "io"

type Encryption interface {

	// Encrypt reads the stream from the provided io.Reader, encrypts it using the
	// provided password and salt and returns the calculated hash of the password and
	// another io.Reader for the encrypted value.
	// The password must be between 4 and 128 (inclusive) characters long.
	Encrypt(value io.Reader, password, salt string) (hash string, encrypted io.Reader, err error)

	// Decrypt takes the same password and salt used when using Encrypt, and decrypts the
	// provided io.Reader stream.
	// See also: Encrypt
	Decrypt(value io.Reader, password, salt string) (io.Reader, error)

	// Verify takes a password and salt, compares the resulting hash to the hash provided
	// to the function, and returns true if they match.
	// See also: Encrypt
	Verify(password, salt, hash string) bool
}

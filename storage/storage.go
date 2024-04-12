package storage

import (
	"io"
	"time"
)

type StoredValue struct {
	io.Reader
	io.Closer
	Name string
}

type StoredValueInfo struct {
	RequiresPassword bool
	Name             string
	Bytes            int64
}

type Storage interface {
	io.Closer

	// Store stores the value encrypted for the given time.Duration after which the value
	// will automatically be deleted.
	// A password is optional (can be set to an empty string), in this case the value will
	// be encrypted using a randomly generated one. This is done to ensure no unencrypted
	// personal data is stored.
	// Returns a unique identifier which can be used to receive the stored value, or an
	// error when the value could not be stored.
	Store(name string, value io.Reader, password string, duration time.Duration) (string, error)

	// Get receives the stored encrypted value by the identifier returned from Store.
	// Returns an error if the given key could not be found or read, or if the given
	// password is incorrect.
	Get(key string, password string) (*StoredValue, error)

	// Remove removes the value from the store, and frees any underlying resources. This
	// operation is idempotent.
	Remove(key string)

	// Info returns any information available for the given key, or nil if there is no
	// information available.
	Info(key string) *StoredValueInfo

	// Size returns the size of the storage in bytes.
	Size() int64
}

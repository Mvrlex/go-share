package storage

import (
	rand2 "crypto/rand"
	"errors"
	"io"
	"io/fs"
	"log"
	"majo-tech.com/share/encryption"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

type StoredFile struct {
	Name              string
	SelfDestructAt    time.Time
	SelfDestruct      *time.Timer
	Bytes             int64
	HasCustomPassword bool
	PwHash            string
}

type FilesystemStorage struct {
	Closed     bool
	basePath   string
	encryption encryption.Encryption
	// Password used when no password is provided to ensure minimally secure encrypted storage.
	fallbackPassword string
	files            map[string]*StoredFile
}

// NewPhysicalStorage initializes file storage in the specified directory. If the
// directory already exists it will delete the directory and all its files, so make sure
// it is not in use.
// Files are always stored encrypted, even when no custom password is provided.
func NewPhysicalStorage(path string) (*FilesystemStorage, error) {

	fallbackPassword, err := generateUid()
	if err != nil {
		return nil, err
	}
	err = createDirectory(path)
	if err == nil {
		return &FilesystemStorage{
			basePath:         path,
			encryption:       &encryption.Aes{},
			fallbackPassword: fallbackPassword,
			files:            make(map[string]*StoredFile),
		}, nil
	}

	if !errors.Is(err, fs.ErrExist) {
		return nil, err
	}
	err = deleteDirectory(path)
	if err != nil {
		return nil, err
	}
	err = createDirectory(path)
	if err != nil {
		return nil, err
	}

	return &FilesystemStorage{
		basePath:         path,
		encryption:       &encryption.Aes{},
		fallbackPassword: fallbackPassword,
		files:            make(map[string]*StoredFile),
	}, nil
}

func (fs *FilesystemStorage) Store(name string, value io.Reader, password string, duration time.Duration) (string, error) {

	key, err := generateUid()
	if err != nil {
		log.Println("could not serve incoming storage request, key generation failed:", err)
		return "", err
	}

	log.Printf("attempting storage of new file with key %q", key)

	effectivePassword := fs.passwordOrFallback(password)
	hash, encrypted, err := fs.encryption.Encrypt(value, effectivePassword, key)
	if err != nil {
		return "", err
	}

	filePath := filepath.Join(fs.basePath, key)
	bytes, err := writeFile(filePath, encrypted)
	if err != nil {
		return "", err
	}

	availableUntil := time.Now().Add(duration)
	fs.files[key] = &StoredFile{
		Name:           name,
		SelfDestructAt: availableUntil,
		SelfDestruct: time.AfterFunc(duration, func() {
			log.Printf("file with key %q reached its storage duration limit, and will now be deleted", key)
			fs.Remove(key)
		}),
		Bytes:             bytes,
		HasCustomPassword: password != "",
		PwHash:            hash,
	}

	log.Printf("new file with key %q stored and will be available until %02d:%02d:%02d", key, availableUntil.Hour(), availableUntil.Minute(), availableUntil.Second())
	log.Printf("now serving %d files with a total size of %d bytes", len(fs.files), fs.Size())
	return key, nil
}

func (fs *FilesystemStorage) Get(key string, password string) (*StoredValue, error) {

	file := fs.files[key]
	if file == nil {
		return nil, &FileNotFoundError{Key: key}
	}
	if file.HasCustomPassword && password == "" {
		return nil, PasswordMissingError
	}

	password = fs.passwordOrFallback(password)
	if !fs.encryption.Verify(password, key, file.PwHash) {
		return nil, PasswordWrongError
	}

	filePath := filepath.Join(fs.basePath, key)
	openedFile, err := os.Open(filePath)
	if err != nil {
		return nil, errors.Join(FileReadError, err)
	}

	decryptReader, err := fs.encryption.Decrypt(openedFile, password, key)
	if err != nil {
		return nil, err
	}

	storedValue := &StoredValue{
		Reader: decryptReader,
		Closer: openedFile,
		Name:   file.Name,
	}

	runtime.SetFinalizer(openedFile, func(*os.File) {
		runtime.KeepAlive(file) // keep file alive for as long as it is read from, see .Remove function below
	})

	return storedValue, nil
}

func (fs *FilesystemStorage) Remove(key string) {
	file := fs.files[key]
	if file == nil {
		return
	}
	file.SelfDestruct.Stop()
	delete(fs.files, key)
	// use finalizer as to not disturb other threads that may currently be reading this file
	runtime.SetFinalizer(file, func(*StoredFile) {
		filePath := filepath.Join(fs.basePath, key)
		if err := os.Remove(filePath); !errors.Is(err, os.ErrNotExist) && err != nil {
			// Design note: Panicking might be a bit overkill here, but this error is
			// not allowed to happen, because it could lead to the application taking more
			// disk space than it is allowed to allocate. Let's just hope the keep alive
			// works as intended, and the file is closed when this finalizer runs ;)
			log.Panicf("could not delete file with key %q at %q from disk: %s", key, fs.basePath, err)
		}
		log.Printf("file with key %q was deleted", key)
		log.Printf("now serving %d files with a total size of %d bytes", len(fs.files), fs.Size())
	})
}

func (fs *FilesystemStorage) Info(key string) *StoredValueInfo {
	storedFile := fs.files[key]
	if storedFile == nil {
		return nil
	}
	return &StoredValueInfo{
		RequiresPassword: storedFile.HasCustomPassword,
		Bytes:            storedFile.Bytes,
		Name:             storedFile.Name,
	}
}

func (fs *FilesystemStorage) Size() int64 {
	var size int64 = 0
	for _, value := range fs.files {
		size += value.Bytes
	}
	return size
}

func (fs *FilesystemStorage) Close() error {
	if fs.Closed {
		return nil
	}
	log.Printf("closing storage and removing directory %q with %d leftover files", fs.basePath, len(fs.files))
	fs.Closed = true
	for _, file := range fs.files {
		file.SelfDestruct.Stop()
	}
	clear(fs.files)
	return deleteDirectory(fs.basePath)
}

func (fs *FilesystemStorage) passwordOrFallback(password string) string {
	if password == "" {
		return fs.fallbackPassword
	}
	return password
}

func writeFile(path string, value io.Reader) (int64, error) {

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return 0, errors.Join(FileCreationError, err)
	}

	bytes, err := io.Copy(file, value)
	if err != nil {
		file.Close()
		os.Remove(path)
		return 0, errors.Join(FileWriteError, err)
	}
	err = file.Close()
	if err != nil {
		os.Remove(path)
		return 0, errors.Join(FileWriteError, err)
	}

	return bytes, nil
}

func createDirectory(path string) error {
	if err := os.Mkdir(path, 0600); err != nil {
		return errors.Join(&DirectoryCreationError{Path: path}, err)
	}
	return nil
}

func deleteDirectory(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return errors.Join(&DirectoryDeletionError{Path: path}, err)
	}
	return nil
}

func generateUid() (string, error) {
	random, err := rand2.Int(rand2.Reader, big.NewInt(int64(math.MaxInt64)))
	if err != nil {
		return "", errors.Join(KeyGenerationError, err)
	}
	return strconv.FormatInt(random.Int64(), 36), nil
}

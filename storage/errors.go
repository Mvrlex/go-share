package storage

import (
	"errors"
	"fmt"
)

var (
	KeyGenerationError    = errors.New("could not generate key for storage")
	FileCreationError     = errors.New("could not create file")
	FileWriteError        = errors.New("could not write to file")
	FileReadError         = errors.New("could not read from file")
	PasswordMissingError  = errors.New("the requested file requires a password and none was provided")
	PasswordWrongError    = errors.New("provided password is incorrect")
	PasswordTooShortError = errors.New("provided password is too short")
	PasswordTooLongError  = errors.New("provided password is too long")
)

type DirectoryCreationError struct {
	Path string
}

func (e *DirectoryCreationError) Error() string {
	return fmt.Sprintf("directory %s does not exist and could not be created", e.Path)
}

func (e *DirectoryCreationError) Is(err error) bool {
	var directoryCreationError *DirectoryCreationError
	return errors.As(err, &directoryCreationError)
}

type DirectoryDeletionError struct {
	Path string
}

func (e *DirectoryDeletionError) Error() string {
	return fmt.Sprintf("directory %s could not be removed", e.Path)
}

func (e *DirectoryDeletionError) Is(err error) bool {
	var directoryDeletionError *DirectoryDeletionError
	return errors.As(err, &directoryDeletionError)
}

type FileNotFoundError struct {
	Key string
}

func (e *FileNotFoundError) Error() string {
	return fmt.Sprintf("no file associated to the key %q was found", e.Key)
}

func (e *FileNotFoundError) Is(err error) bool {
	var fileNotFoundError *FileNotFoundError
	return errors.As(err, &fileNotFoundError)
}

package server

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

var (
	MaxBodySizeError = errors.New("reached file upload limit")
)

type UploadRequest struct {
	Duration        time.Duration
	Password        string
	FileName        string
	FileReader      io.Reader
	MultiPartReader *multipart.Reader
}

func ReadUploadFormData(reader *multipart.Reader) (*UploadRequest, error) {

	duration, err := readDuration(reader)
	if err != nil {
		return nil, err
	}

	password, err := readPassword(reader)
	if err != nil {
		return nil, err
	}

	fileName, filePartReader, err := getFile(reader)
	if err != nil {
		return nil, err
	}

	return &UploadRequest{
		Duration:        duration,
		Password:        password,
		FileName:        fileName,
		FileReader:      filePartReader,
		MultiPartReader: reader,
	}, nil
}

func readDuration(reader *multipart.Reader) (time.Duration, error) {
	part, err := nextPart(reader, "duration")
	if err != nil {
		return 0, err
	}
	partData, err := io.ReadAll(part)
	if err != nil {
		return 0, errors.Join(errors.New("could not read part"), err)
	}
	duration, err := time.ParseDuration(string(partData))
	if err != nil {
		return 0, errors.Join(errors.New(fmt.Sprintf("duration %q could not be parsed", string(partData))), err)
	}
	if duration.Hours() > 24 || duration.Seconds() < 1 {
		return 0, errors.New("requested duration is outside of valid range")
	}
	return duration, nil
}

func readPassword(reader *multipart.Reader) (string, error) {
	part, err := nextPart(reader, "password")
	if err != nil {
		return "", err
	}
	partData, err := io.ReadAll(part)
	if err != nil {
		return "", errors.Join(errors.New("could not read part"), err)
	}
	if len(partData) != 0 && (len(partData) > 127 || len(partData) < 3) {
		return "", errors.New("password has an invalid length")
	}
	return string(partData), nil
}

func getFile(reader *multipart.Reader) (fileName string, filePartReader io.Reader, err error) {
	part, err := nextPart(reader, "file")
	if err != nil {
		return "", nil, err
	}
	return part.FileName(), part, nil
}

func nextPart(reader *multipart.Reader, expectedFormName string) (*multipart.Part, error) {
	part, err := reader.NextPart()
	if err != nil {
		var bytesError *http.MaxBytesError
		if errors.As(err, &bytesError) {
			return nil, errors.Join(MaxBodySizeError, err)
		}
		return nil, errors.Join(errors.New(fmt.Sprintf("could not read next part %q", expectedFormName)), err)
	}
	if part.FormName() != expectedFormName {
		return nil, errors.New(fmt.Sprintf("expected part %q but got %q", expectedFormName, part.FormName()))
	}
	return part, nil
}

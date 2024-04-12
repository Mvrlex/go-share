package server

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/skip2/go-qrcode"
	"io"
	"log"
	"majo-tech.com/share/storage"
	"majo-tech.com/share/templates"
	"net/http"
)

type UploadHandler struct {
	Storage          storage.Storage
	Templates        templates.Templates
	MaxFileSizeBytes int64
	DiskSpaceBytes   int64
	Host             string
}

func (u *UploadHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {

	// Ok, so apparently none of the big browsers support this part of the HTTP spec:
	// > An HTTP/1.1 (or later) client sending a message-body SHOULD monitor the network
	// > connection for an error status while it is transmitting the request. If the
	// > client sees an error status, it SHOULD immediately cease transmitting the body.
	// When a browser sends a file, and we respond with an error, it simply ignores it
	// and keeps sending the contents. This will force Go to break the connection. The
	// browser will then retry the request 3 more times until it finally gets the fucking
	// hint.
	// Unfortunately the error we sent will then be lost forever, and the browser simply
	// fails with an ERR_CONNECTION_RESET error.
	// We would have to read the full request, consuming resources, making us vulnerable
	// to slow DOS attacks, and forcing the user to upload a file, even though the request
	// will never work.
	// The better option is probably to let Go break the connection, and report a generic
	// error to the user.
	// Figuring this shit out was such a fucking pain... not even IntelliJ's HTTP client
	// does this correctly, so initially I thought this is a bug on Go's side. I'm sorry
	// Go, for doubting you.
	// defer io.Copy(io.Discard, request.Body)

	maxFileSize := u.MaxFileSizeBytes + 1048576 // add 1 MB for the other form values
	if request.ContentLength >= maxFileSize || request.ContentLength == -1 {
		log.Println("request rejected because request body is too large")
		WriteError(writer, u.Templates, http.StatusRequestEntityTooLarge, "Your file exceeds the maximum allowed size for file uploads.")
		return
	}

	if u.Storage.Size()+request.ContentLength > u.DiskSpaceBytes { // 30 GB
		log.Println("request rejected because it would exceed the servers maximum allowed capacity")
		WriteError(writer, u.Templates, http.StatusBadRequest, "Your file would exceed our servers current maximum capacity, please try again later.")
		return
	}

	request.Body = http.MaxBytesReader(writer, request.Body, maxFileSize)
	multipartReader, err := request.MultipartReader()
	if err != nil {
		log.Println("could not create multipart reader:", err)
		WriteError(writer, u.Templates, http.StatusInternalServerError, "Your request was malformed and our server could not handle it, this seems to be a problem on our side.")
		return
	}

	uploadRequest, err := ReadUploadFormData(multipartReader)
	if err != nil {
		if errors.Is(err, MaxBodySizeError) {
			log.Println("request rejected because request body is too large")
			WriteError(writer, u.Templates, http.StatusBadRequest, "Your file exceeds the maximum allowed size for file uploads.")
			return
		}
		log.Println("malformed multipart request received:", err)
		WriteError(writer, u.Templates, http.StatusInternalServerError, "Your request was malformed and our server could not handle it, this seems to be a problem on our side.")
		return
	}

	key, err := u.Storage.Store(uploadRequest.FileName, uploadRequest.FileReader, uploadRequest.Password, uploadRequest.Duration)
	if err != nil {
		log.Println("storing file failed:", err)
		WriteError(writer, u.Templates, http.StatusInternalServerError, "We could not store your file on our server, this seems to be a problem on our side.")
		return
	}

	part, err := uploadRequest.MultiPartReader.NextPart()
	if err != io.EOF {
		log.Println("additional content after file: ", part.FormName())
		u.Storage.Remove(key)
		WriteError(writer, u.Templates, http.StatusBadRequest, "Unexpected additional content after file. Uploading multiple files is not allowed.")
		return
	}

	u.respondWithLinkToDownloadFile(writer, key, u.Host)
}

func (u *UploadHandler) respondWithLinkToDownloadFile(writer http.ResponseWriter, key string, host string) {
	link := fmt.Sprintf("%s/shares/%s", host, key)
	code, err := generateQrCode(link)
	if err != nil {
		log.Println("generating qr-code failed, qr-code will be omitted from response:", err)
	}
	writer.Header().Add("Content-Type", "text/html")
	err = u.Templates.TemplateUploadSuccess(writer, code, link)
	if err != nil {
		u.Storage.Remove(key) // sad... all that work for nothing ;(
		log.Println("could not respond with share template", err)
		WriteError(writer, u.Templates, http.StatusInternalServerError, "Storing the file was successful, but we could not serve you a response page. We deleted your file, please try again.")
		return
	}
}

func generateQrCode(data string) ([]byte, error) {
	png, err := qrcode.Encode(data, qrcode.Medium, 256)
	if err != nil {
		return nil, err
	}
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(png)))
	base64.StdEncoding.Encode(dst, png)
	return dst, nil
}

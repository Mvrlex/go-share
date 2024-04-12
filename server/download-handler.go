package server

import (
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"log"
	"majo-tech.com/share/storage"
	"majo-tech.com/share/templates"
	"net/http"
)

type DownloadHandler struct {
	Storage   storage.Storage
	Templates templates.Templates
}

func (u *DownloadHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {

	request.Body = http.MaxBytesReader(writer, request.Body, 1024)
	err := request.ParseForm()
	if err != nil && err != io.EOF {
		log.Println("could not parse body of download request:", err)
		WriteErrorPage(writer, u.Templates, http.StatusBadRequest, "Your request was malformed.")
		return
	}
	password := request.PostFormValue("password")

	key := mux.Vars(request)["key"]
	storedFile, err := u.Storage.Get(key, password)
	if err != nil {
		if errors.Is(err, &storage.FileNotFoundError{}) {
			WriteErrorPage(writer, u.Templates, http.StatusNotFound, "This file (no longer?) exists.")
		} else if errors.Is(err, storage.PasswordWrongError) {
			WriteErrorPage(writer, u.Templates, http.StatusBadRequest, "The provided password is incorrect.")
		} else {
			log.Println("download request from user caused an error:", err)
			WriteErrorPage(writer, u.Templates, http.StatusBadRequest, "Your request was malformed.")
		}
		return
	}
	done := false
	defer func() {
		if done {
			u.Storage.Remove(key)
		}
	}()
	defer storedFile.Close()

	writer.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", storedFile.Name))
	_, err = io.Copy(writer, storedFile)
	if err != nil {
		log.Println("could not send file to user:", err)
		WriteErrorPage(writer, u.Templates, http.StatusInternalServerError, "Your request was malformed.")
		return
	}

	log.Printf("successfully downloaded file with key %q", key)
	done = true

}

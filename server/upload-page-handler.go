package server

import (
	"log"
	"majo-tech.com/share/storage"
	"majo-tech.com/share/templates"
	"net/http"
)

type UploadPageHandler struct {
	Storage          storage.Storage
	Templates        templates.Templates
	MaxFileSizeBytes int64
}

func (u *UploadPageHandler) ServeHTTP(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Add("Content-Type", "text/html")
	err := u.Templates.TemplateUpload(writer, u.MaxFileSizeBytes, u.Storage.Size())
	if err != nil {
		log.Println("could not respond with upload page,", err)
		writer.WriteHeader(500)
		return
	}
	log.Println("serving upload page")
}

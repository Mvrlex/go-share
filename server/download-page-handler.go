package server

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"majo-tech.com/share/storage"
	"majo-tech.com/share/templates"
	"net/http"
)

type DownloadPageHandler struct {
	Storage   storage.Storage
	Templates templates.Templates
}

func (u *DownloadPageHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {

	key := mux.Vars(request)["key"]

	info := u.Storage.Info(key)
	if info == nil {
		WriteErrorPage(writer, u.Templates, 404, "This file (no longer?) exists.")
		return
	}

	writer.Header().Add("Content-Type", "text/html")
	err := u.Templates.TemplateDownload(writer, key, info.Name, ByteCountIEC(info.Bytes), info.RequiresPassword)
	if err != nil {
		log.Println("could not respond with download template", err)
		WriteError(writer, u.Templates, http.StatusInternalServerError, "Something majorly broke, and we could not serve you a response page.")
		return
	}

}

func ByteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}

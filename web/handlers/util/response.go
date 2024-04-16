package util

import (
	"majo-tech.com/share/web/templates"
	"net/http"
)

func WriteError(writer http.ResponseWriter, templates templates.Templates, status int, desc string) {
	writer.Header().Add("Content-Type", "text/html; charset=utf-8")
	writer.WriteHeader(status)
	err := templates.TemplateError(writer, desc)
	if err != nil {
		http.Error(writer, desc, status)
	}
}

func WriteErrorPage(writer http.ResponseWriter, templates templates.Templates, status int, desc string) {
	writer.Header().Add("Content-Type", "text/html; charset=utf-8")
	writer.WriteHeader(status)
	err := templates.TemplateError(writer, desc)
	if err != nil {
		http.Error(writer, desc, status)
	}
}
